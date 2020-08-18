// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"errors"
	"fmt"
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
	Logger           *Logger          `bson:"keyhole"`
	OplogStats       OplogStats       `bson:"oplog"`
	Process          string           `bson:"process"`
	ReplSetGetStatus ReplSetGetStatus `bson:"replSetGetStatus"`
	ServerStatus     ServerStatus     `bson:"serverStatus"`
	Shards           []Shard          `bson:"shards"`
	Version          string           `bson:"version"`

	redact  bool
	verbose bool
}

// NewStats -
func NewStats(version string) *ClusterStats {
	s := ClusterStats{Logger: NewLogger(version, "-allinfo")}
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
	p.Logger.Log("GetClusterStats() begins")
	if err = p.GetClusterStatsSummary(client); err != nil {
		return err
	}
	if p.CmdLineOpts, err = GetCmdLineOpts(client); err != nil {
		p.Logger.Log(fmt.Sprintf(`GetCmdLineOpts(): %v`, err))
	}
	if p.Cluster == Sharded { //collects from the primary of each shard
		p.Logger.Log("sharded detected, collecting from all servers")
		if p.Shards, err = GetShards(client); err != nil {
			p.Logger.Log(fmt.Sprintf(`GetShards(): %v`, err))
		}
		if p.Shards, err = p.GetServersStatsSummary(p.Shards, connString); err != nil {
			p.Logger.Log(fmt.Sprintf(`GeterversStatsSummary(): %v`, err))
		}
		p.Logger.Log("end collecting from all servers")
	} else if p.Cluster == Replica && p.Process == "mongod" { //collects replica info
		message := "replica detected, collecting from all servers"
		p.Logger.Log(message)
		if p.ReplSetGetStatus, err = GetReplSetGetStatus(client); err != nil {
			p.Logger.Log(fmt.Sprintf(`GetReplSetGetStatus(): %v`, err))
		}

		setName := p.ServerStatus.Repl.SetName
		s := fmt.Sprintf(`%v/%v`, setName, strings.Join(p.ServerStatus.Repl.Hosts, ","))
		oneShard := []Shard{Shard{ID: setName, State: 1, Host: s}}
		if p.Shards, err = p.GetServersStatsSummary(oneShard, connString); err != nil {
			p.Logger.Log(fmt.Sprintf(`GeterversStatsSummary(): %v`, err))
		}
		p.Logger.Log("end collecting from all servers")
	}
	db := NewDatabaseStats(p.Logger.Version)
	db.SetRedaction(p.redact)
	db.SetVerbose(p.verbose)
	var databases []Database
	if databases, err = db.GetAllDatabasesStats(client); err != nil {
		p.Logger.Log(fmt.Sprintf(`GetAllDatabasesStats(): %v`, err))
	}
	for _, m := range db.GetLogs() {
		p.Logger.Add(m)
	}
	p.Databases = databases
	return nil
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
	outdir := "./out/"
	os.Mkdir(outdir, 0755)
	ofile := outdir + p.HostInfo.System.Hostname + "-stats.bson.gz"
	if err = gox.OutputGzipped(data, ofile); err != nil {
		return err
	}
	fmt.Println(fmt.Sprintf(`bson data written to %v`, ofile))
	return err
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
			p.Logger.Log(fmt.Sprintf(`GetOplogStats(): %v`, err))
		}
	}
	return nil
}

// GetClusterShortSummary returns one line summary
func (p *ClusterStats) GetClusterShortSummary(client *mongo.Client) string {
	var err error
	if err = p.GetClusterStatsSummary(client); err != nil {
		return err.Error()
	}
	return p.PrintShortSummary()
}

// GetServersStatsSummary returns cluster stats from all shards
func (p *ClusterStats) GetServersStatsSummary(shards []Shard, connString connstring.ConnString) ([]Shard, error) {
	var err error
	var uris []string
	var smap = map[string]Shard{}
	for _, v := range shards {
		v.Servers = []ClusterStats{}
		smap[v.ID] = v
	}
	if uris, err = GetAllServerURIs(shards, connString); err != nil {
		p.Logger.Log(err.Error())
		return shards, err
	}
	wg := gox.NewWaitGroup(4)
	var mu sync.Mutex
	for i, uri := range uris {
		s := uri
		cs, _ := connstring.Parse(s)
		if cs.Password != "" {
			s = strings.ReplaceAll(s, cs.Password, "xxxxxx")
		}
		msg := fmt.Sprintf(`[t-%d] begin collecting from %v`, i, s)
		p.Logger.Log(msg)
		msg = fmt.Sprintf(`[t-%d] end collecting from %v`, i, s)
		wg.Add(1)
		go func(uri string, msg string) {
			defer wg.Done()
			defer p.Logger.Log(msg)
			var sclient *mongo.Client
			if sclient, err = NewMongoClient(uri, connString.SSLCaFile, connString.SSLClientCertificateKeyFile); err != nil {
				p.Logger.Log(err.Error())
				return
			}
			defer sclient.Disconnect(context.Background())
			server := NewStats(p.Logger.Version)
			if err = server.GetClusterStatsSummary(sclient); err != nil {
				p.Logger.Log(err.Error())
				return
			}
			mu.Lock()
			node := smap[server.ServerStatus.Repl.SetName]
			node.Servers = append(node.Servers, *server)
			smap[server.ServerStatus.Repl.SetName] = node
			mu.Unlock()
		}(uri, msg)
	}
	wg.Wait()

	shards = []Shard{}
	for _, v := range smap {
		shards = append(shards, v)
	}
	return shards, nil
}

// PrintShortSummary prints a short summary
func (p *ClusterStats) PrintShortSummary() string {
	edition := "community"
	if len(p.BuildInfo.Modules) > 0 {
		edition = p.BuildInfo.Modules[0]
	}
	result := fmt.Sprintf(`MongoDB v%v %v %v (%v) %v %v %v cores %v mem`,
		p.BuildInfo.Version, edition, p.HostInfo.System.Hostname, p.HostInfo.OS.Name,
		p.ServerStatus.Process, p.Cluster, p.HostInfo.System.NumCores, p.HostInfo.System.MemSizeMB)
	return result
}
