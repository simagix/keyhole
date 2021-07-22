// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"

	"github.com/simagix/gox"
	"go.mongodb.org/mongo-driver/mongo"
)

// Sharded cluster
const Sharded = "sharded"

// Replica set
const Replica = "replica"

// Standalone server
const Standalone = "standalone"

// ClusterStats keeps slow ops struct
type ClusterStats struct {
	BuildInfo        BuildInfo        `bson:"buildInfo"`
	CmdLineOpts      CmdLineOpts      `bson:"getCmdLineOpts"`
	Cluster          string           `bson:"cluster"`
	Databases        []Database       `bson:"databases"`
	Host             string           `bson:"host"`
	HostInfo         HostInfo         `bson:"hostInfo"`
	Logger           *gox.Logger      `bson:"keyhole"`
	OplogStats       OplogStats       `bson:"oplog"`
	Process          string           `bson:"process"`
	ReplSetGetStatus ReplSetGetStatus `bson:"replSetGetStatus"`
	ServerStatus     ServerStatus     `bson:"serverStatus"`
	Shards           []Shard          `bson:"shards"`
	Version          string           `bson:"version"`

	redact  bool
	verbose bool
	version string
}

// NewClusterStats returns *ClusterStats
func NewClusterStats(version string) *ClusterStats {
	s := ClusterStats{version: version}
	return &s
}

// SetRedaction sets redact
func (p *ClusterStats) SetRedaction(redact bool) {
	p.redact = redact
}

// SetVerbose sets verbose mode
func (p *ClusterStats) SetVerbose(verbose bool) {
	p.verbose = verbose
}

// GetClusterStats collects cluster stats
func (p *ClusterStats) GetClusterStats(client *mongo.Client, connString connstring.ConnString) error {
	var err error
	if p.Logger == nil {
		p.Logger = gox.GetLogger(p.version)
		if p.verbose {
			p.Logger.SetLoggerLevel(gox.Debug)
		}
	}
	p.Logger.Info("GetClusterStats() begins")
	if err = p.GetClusterStatsSummary(client); err != nil {
		return err
	}
	if p.CmdLineOpts, err = GetCmdLineOpts(client); err != nil {
		p.Logger.Info(fmt.Sprintf(`GetCmdLineOpts(): %v`, err))
	}
	if p.Cluster == Sharded { //collects from the primary of each shard
		if p.Shards, err = GetShards(client); err != nil {
			p.Logger.Infof(`GetShards(): %v`, err)
		}
		p.Logger.Infof("%v shards detected, collecting from all servers", len(p.Shards))
		if p.Shards, err = p.GetServersStatsSummary(p.Shards, connString); err != nil {
			p.Logger.Error(err)
		}
		p.Logger.Info("end collecting from all servers")
	} else if p.Cluster == Replica && p.Process == "mongod" { //collects replica info
		message := "replica detected, collecting from all servers"
		p.Logger.Info(message)
		if p.ReplSetGetStatus, err = GetReplSetGetStatus(client); err != nil {
			p.Logger.Info(fmt.Sprintf(`GetReplSetGetStatus(): %v`, err))
		}

		setName := p.ServerStatus.Repl.SetName
		s := fmt.Sprintf(`%v/%v`, setName, strings.Join(p.ServerStatus.Repl.Hosts, ","))
		oneShard := []Shard{Shard{ID: setName, State: 1, Host: s}}
		if p.Shards, err = p.GetServersStatsSummary(oneShard, connString); err != nil {
			p.Logger.Error(err)
		}
		p.Logger.Info("end collecting from all servers")
	}
	db := NewDatabaseStats(p.Logger.AppName)
	db.SetRedaction(p.redact)
	db.SetVerbose(p.verbose)
	if p.Databases, err = db.GetAllDatabasesStats(client); err != nil {
		p.Logger.Info(fmt.Sprintf(`GetAllDatabasesStats(): %v`, err))
	}
	return nil
}

// GetClusterStatsSummary collects cluster stats
func (p *ClusterStats) GetClusterStatsSummary(client *mongo.Client) error {
	var err error
	if p.BuildInfo, err = GetBuildInfo(client); err != nil {
		return err
	}
	p.Version = p.BuildInfo.Version
	if p.HostInfo, err = GetHostInfo(client); err != nil {
		return err
	}
	if p.ServerStatus, err = GetServerStatus(client); err != nil {
		return err
	}
	p.Host = p.ServerStatus.Host
	p.Process = p.ServerStatus.Process
	p.Cluster = GetClusterType(p.ServerStatus)
	if p.Cluster == Replica && p.Process == "mongod" { //collects replica info
		if p.OplogStats, err = GetOplogStats(client); err != nil {
			return err
		}
		if p.ReplSetGetStatus, err = GetReplSetGetStatus(client); err != nil {
			return err
		}
	} else if p.Cluster == Sharded {
		if p.Shards, err = GetShards(client); err != nil {
			return err
		}
	}
	return nil
}

// GetServersStatsSummary returns cluster stats from all shards
func (p *ClusterStats) GetServersStatsSummary(shards []Shard, connString connstring.ConnString) ([]Shard, error) {
	var err error
	var uris []string
	var smap = map[string]Shard{}
	if p.Logger == nil {
		p.Logger = gox.GetLogger(p.version)
	}
	for _, v := range shards {
		v.Servers = []ClusterStats{}
		smap[v.ID] = v
	}
	if uris, err = GetAllServerURIs(shards, connString); err != nil {
		return shards, err
	}
	wg := gox.NewWaitGroup(6)
	var mu sync.Mutex
	var e error
	for i, uri := range uris {
		s := uri
		cs, _ := connstring.Parse(s)
		if cs.Password != "" {
			s = strings.Replace(s, url.QueryEscape(cs.Password), "xxxxxx", 1)
		}
		p.Logger.Debugf(`[t-%d] begin collecting from %v`, i, s)
		wg.Add(1)
		go func(uri string, n int, logger *gox.Logger) {
			defer wg.Done()
			var sclient *mongo.Client
			var err error
			if sclient, err = NewMongoClient(uri); err != nil {
				logger.Infof(`[t-%d] error collecting: %v`, n, err)
				mu.Lock()
				e = err
				mu.Unlock()
				return
			}
			defer sclient.Disconnect(context.Background())
			server := NewClusterStats(p.Logger.AppName)
			if e = server.GetClusterStatsSummary(sclient); e != nil {
				logger.Infof(`[t-%d] error collecting: %v`, n, err)
				mu.Lock()
				e = err
				mu.Unlock()
				return
			}
			mu.Lock()
			node := smap[server.ServerStatus.Repl.SetName]
			node.Servers = append(node.Servers, *server)
			smap[server.ServerStatus.Repl.SetName] = node
			mu.Unlock()
			logger.Debugf(`[t-%d] end collecting from %v`, n, s)
		}(uri, i, p.Logger)
	}
	wg.Wait()
	if e != nil {
		return shards, e
	}
	shards = []Shard{}
	for _, v := range smap {
		shards = append(shards, v)
	}
	return shards, nil
}

// GetClusterShortSummary returns one line summary
func (p *ClusterStats) GetClusterShortSummary(client *mongo.Client) string {
	var err error
	if err = p.GetClusterStatsSummary(client); err != nil {
		return err.Error()
	}
	return p.GetShortSummary()
}

// GetShortSummary returns a short summary
func (p *ClusterStats) GetShortSummary() string {
	edition := "community"
	if len(p.BuildInfo.Modules) > 0 {
		edition = p.BuildInfo.Modules[0]
	}
	numShardStr := ""
	if p.Cluster == Sharded {
		numShardStr = fmt.Sprintf(`(%v)`, len(p.Shards))
	}
	result := fmt.Sprintf(`MongoDB v%v %v %v (%v) %v %v%v %v cores %v mem`,
		p.BuildInfo.Version, edition, p.HostInfo.System.Hostname, p.HostInfo.OS.Name,
		p.ServerStatus.Process, p.Cluster, numShardStr, p.HostInfo.System.NumCores, p.HostInfo.System.MemSizeMB)
	return result
}

// Print prints a cluster short summary
func (p *ClusterStats) Print() {
	fmt.Println(p.GetShortSummary())
}

// OutputBSON writes bson data to a file
func (p *ClusterStats) OutputBSON() error {
	if p.HostInfo.System.Hostname == "" {
		result := `Roles 'clusterMonitor' and 'readAnyDatabase' are required`
		return errors.New(result)
	}
	var err error
	var data []byte
	if data, err = bson.Marshal(p); err != nil {
		return err
	}

	outdir := "./out"
	os.Mkdir(outdir, 0755)
	basename := p.HostInfo.System.Hostname
	basename = strings.ReplaceAll(basename, ":", "_")
	ofile := fmt.Sprintf(`%v/%v-stats.bson.gz`, outdir, basename)
	i := 1
	for DoesFileExist(ofile) {
		ofile = fmt.Sprintf(`%v/%v.%d-stats.bson.gz`, outdir, basename, i)
		i++
	}

	if err = gox.OutputGzipped(data, ofile); err != nil {
		return err
	}
	fmt.Println(fmt.Sprintf(`bson data written to %v`, ofile))
	return err
}
