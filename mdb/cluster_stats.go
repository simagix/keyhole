// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

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
	Databases        *[]Database      `bson:"databases"`
	Host             string           `bson:"host"`
	HostInfo         HostInfo         `bson:"hostInfo"`
	Logger           *gox.Logger      `bson:"keyhole"`
	OplogStats       OplogStats       `bson:"oplog"`
	Process          string           `bson:"process"`
	ReplSetGetStatus ReplSetGetStatus `bson:"replSetGetStatus"`
	ServerStatus     ServerStatus     `bson:"serverStatus"`
	Shards           []Shard          `bson:"shards"`
	Version          string           `bson:"version"`

	dbNames   []string
	fastMode  bool
	redact    bool
	signature string
	verbose   bool
}

// NewClusterStats returns *ClusterStats
func NewClusterStats(signature string) *ClusterStats {
	s := ClusterStats{signature: signature}
	return &s
}

// SetDBNames sets redact
func (p *ClusterStats) SetDBNames(dbNames []string) {
	p.dbNames = dbNames
}

// SetFastMode sets fastMode mode
func (p *ClusterStats) SetFastMode(fastMode bool) {
	p.fastMode = fastMode
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
	p.Logger = gox.GetLogger(p.signature)
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
		hosts := []string{}
		for _, member := range p.ReplSetGetStatus.Members {
			hosts = append(hosts, member.Name)
		}
		s := fmt.Sprintf(`%v/%v`, setName, strings.Join(hosts, ","))
		oneShard := []Shard{{ID: setName, State: 1, Host: s}}
		if p.Shards, err = p.GetServersStatsSummary(oneShard, connString); err != nil {
			p.Logger.Error(err)
		}
		p.Logger.Info("end collecting from all servers")
	}
	db := NewDatabaseStats(p.Logger.AppName)
	db.SetNumberShards(len(p.Shards))
	db.SetRedaction(p.redact)
	db.SetVerbose(p.verbose)
	db.SetFastMode(p.fastMode)
	var databases []Database
	if databases, err = db.GetAllDatabasesStats(client, p.dbNames); err != nil {
		p.Logger.Info(fmt.Sprintf(`GetAllDatabasesStats(): %v`, err))
	}
	p.Databases = &databases
	return nil
}

// GetClusterStatsSummary collects cluster stats
func (p *ClusterStats) GetClusterStatsSummary(client *mongo.Client) error {
	var err error
	p.Logger = gox.GetLogger(p.signature)
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
	p.Logger = gox.GetLogger(p.signature)
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
		p.Logger.Infof(`[t-%d] collect from %v`, i, s)
		wg.Add(1)
		go func(uri string, n int, logger *gox.Logger) {
			defer wg.Done()
			var sclient *mongo.Client
			var err error
			if sclient, err = NewMongoClient(uri); err != nil {
				logger.Errorf(`[t-%d] error: %v`, n, err)
				mu.Lock()
				e = err
				mu.Unlock()
				return
			}
			tm := time.Now()
			sclient.Ping(context.Background(), nil)
			logger.Infof(`[t-%d] ping: %v`, n, time.Since(tm))
			defer sclient.Disconnect(context.Background())
			server := NewClusterStats(p.Logger.AppName)
			if err = server.GetClusterStatsSummary(sclient); err != nil {
				logger.Errorf(`[t-%d] error: %v`, n, err)
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
			logger.Infof(`[t-%d] completed`, n)
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
func (p *ClusterStats) OutputBSON() (string, []byte, error) {
	var err error
	var data []byte
	var ofile string
	if p.HostInfo.System.Hostname == "" {
		result := `roles 'clusterMonitor' and 'readAnyDatabase' are required`
		return ofile, data, errors.New(result)
	}

	os.Mkdir(outdir, 0755)
	basename := p.HostInfo.System.Hostname
	basename = strings.ReplaceAll(basename, ":", "_")
	ofile = fmt.Sprintf(`%v/%v-stats.bson.gz`, outdir, basename)
	i := 1
	for DoesFileExist(ofile) {
		ofile = fmt.Sprintf(`%v/%v.%d-stats.bson.gz`, outdir, basename, i)
		i++
	}

	databases := p.Databases
	p.Databases = nil
	var summaries []Database
	for _, db := range *databases {
		dbSummary := Database{
			Name:       db.Name,
			SizeOnDisk: db.SizeOnDisk,
			Empty:      db.Empty,
			Shards:     db.Shards,
			Stats:      db.Stats}
		summaries = append(summaries, dbSummary)
	}
	p.Databases = &summaries
	var buffer bytes.Buffer
	if data, err = bson.Marshal(p); err != nil {
		return ofile, data, err
	}
	nw := 0
	var n int
	for nw < len(data) {
		if n, err = buffer.Write(data); err != nil {
			return ofile, data, err
		}
		nw += n
	}

	for _, db := range *databases {
		for _, coll := range db.Collections {
			if data, err = bson.Marshal(coll); err != nil {
				return ofile, data, err
			}
			nw := 0
			var n int
			for nw < len(data) {
				if n, err = buffer.Write(data); err != nil {
					return ofile, data, err
				}
				nw += n
			}
		}
	}

	if err = gox.OutputGzipped(buffer.Bytes(), ofile); err != nil {
		return ofile, data, err
	}
	fmt.Printf("bson data written to %v\n", ofile)
	return ofile, buffer.Bytes(), err
}
