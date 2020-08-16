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
	bsonExt string
	htmlExt string
	redact  bool
	verbose bool
	version string
}

// ClusterDetails stores cluster details
type ClusterDetails struct {
	BuildInfo        BuildInfo        `bson:"buildInfo"`
	CmdLineOpts      CmdLineOpts      `json:"getCmdLineOpts" bson:"getCmdLineOpts"`
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
}

// NewStats -
func NewStats(version string) *ClusterStats {
	s := ClusterStats{version: version,
		bsonExt: "-stats.bson.gz", htmlExt: "-stats.html"}
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
func (p *ClusterStats) GetClusterStats(client *mongo.Client, connString connstring.ConnString) (ClusterDetails, error) {
	var err error
	var cluster = ClusterDetails{Logger: NewLogger(p.version, "-allinfo")}
	cluster.Logger.Log("GetClusterStats() begins")
	if cluster, err = p.GetClusterStatsSummary(client); err != nil {
		return cluster, err
	}
	if cluster.CmdLineOpts, err = GetCmdLineOpts(client); err != nil {
		cluster.Logger.Log(fmt.Sprintf(`GetCmdLineOpts(): %v`, err))
	}
	if cluster.Cluster == Sharded { //collects from the primary of each shard
		cluster.Logger.Log("sharded detected, collecting from all servers")
		if cluster.Shards, err = GetShards(client); err != nil {
			cluster.Logger.Log(fmt.Sprintf(`GetShards(): %v`, err))
		}
		if cluster.Shards, err = p.GetServersStatsSummary(cluster.Shards, connString, cluster.Logger); err != nil {
			cluster.Logger.Log(fmt.Sprintf(`GeterversStatsSummary(): %v`, err))
		}
		cluster.Logger.Log("end collecting from all servers")
	} else if cluster.Cluster == Replica && cluster.Process == "mongod" { //collects replica info
		message := "replica detected, collecting from all servers"
		cluster.Logger.Log(message)
		if cluster.ReplSetGetStatus, err = GetReplSetGetStatus(client); err != nil {
			cluster.Logger.Log(fmt.Sprintf(`GetReplSetGetStatus(): %v`, err))
		}

		setName := cluster.ServerStatus.Repl.SetName
		s := fmt.Sprintf(`%v/%v`, setName, strings.Join(cluster.ServerStatus.Repl.Hosts, ","))
		oneShard := []Shard{Shard{ID: setName, State: 1, Host: s}}
		if cluster.Shards, err = p.GetServersStatsSummary(oneShard, connString, cluster.Logger); err != nil {
			cluster.Logger.Log(fmt.Sprintf(`GeterversStatsSummary(): %v`, err))
		}
		cluster.Logger.Log("end collecting from all servers")
	}
	db := NewDatabaseStats(p.version)
	db.SetVerbose(p.verbose)
	var databases []Database
	if databases, err = db.GetAllDatabasesStats(client); err != nil {
		cluster.Logger.Log(fmt.Sprintf(`GetAllDatabasesStats(): %v`, err))
	}
	for _, m := range db.GetLogs() {
		cluster.Logger.Add(m)
	}
	cluster.Databases = databases
	return cluster, nil
}

// Save saves to data
func (p *ClusterStats) Save(cluster ClusterDetails) (string, error) {
	if cluster.HostInfo.System.Hostname == "" {
		result := `Roles 'clusterMonitor' and 'readAnyDatabase' are required`
		return "", errors.New(result)
	}
	var err error
	var data []byte
	var results []string
	data, err = bson.Marshal(cluster)
	outdir := "./out/"
	os.Mkdir(outdir, 0755)
	ofile := outdir + cluster.HostInfo.System.Hostname + p.bsonExt
	gox.OutputGzipped(data, ofile)
	results = append(results, fmt.Sprintf(`bson data written to %v`, ofile))
	return strings.Join(results, "\n"), err
}

// GetClusterStatsSummary collects cluster stats
func (p *ClusterStats) GetClusterStatsSummary(client *mongo.Client) (ClusterDetails, error) {
	var err error
	var cluster = ClusterDetails{Logger: NewLogger(p.version, "-allinfo")}
	if cluster.BuildInfo, err = GetBuildInfo(client); err != nil {
		return cluster, err
	}
	cluster.Version = cluster.BuildInfo.Version
	if cluster.HostInfo, err = GetHostInfo(client); err != nil {
		return cluster, err
	}
	if cluster.ServerStatus, err = GetServerStatus(client); err != nil {
		return cluster, err
	}
	cluster.Host = cluster.ServerStatus.Host
	cluster.Process = cluster.ServerStatus.Process
	cluster.Cluster = GetClusterType(cluster.ServerStatus)
	if cluster.Cluster == Replica && cluster.Process == "mongod" { //collects replica info
		if cluster.OplogStats, err = GetOplogStats(client); err != nil {
			cluster.Logger.Log(fmt.Sprintf(`GetOplogStats(): %v`, err))
		}
	}
	return cluster, nil
}

// GetClusterShortSummary returns one line summary
func (p *ClusterStats) GetClusterShortSummary(client *mongo.Client) string {
	var err error
	var c ClusterDetails
	if c, err = p.GetClusterStatsSummary(client); err != nil {
		return err.Error()
	}
	edition := "community"
	if len(c.BuildInfo.Modules) > 0 {
		edition = c.BuildInfo.Modules[0]
	}
	result := fmt.Sprintf(`MongoDB v%v %v %v (%v) %v %v %v cores %v mem`,
		c.BuildInfo.Version, edition, c.HostInfo.System.Hostname, c.HostInfo.OS.Name,
		c.ServerStatus.Process, c.Cluster, c.HostInfo.System.NumCores, c.HostInfo.System.MemSizeMB)
	return result
}

// GetServersStatsSummary returns cluster stats from all shards
func (p *ClusterStats) GetServersStatsSummary(shards []Shard, connString connstring.ConnString, logger *Logger) ([]Shard, error) {
	var err error
	var uris []string
	var smap = map[string]Shard{}
	for _, v := range shards {
		v.Servers = []ClusterDetails{}
		smap[v.ID] = v
	}
	if uris, err = GetAllServerURIs(shards, connString); err != nil {
		logger.Log(err.Error())
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
		logger.Log(msg)
		msg = fmt.Sprintf(`[t-%d] end collecting from %v`, i, s)
		wg.Add(1)
		go func(uri string, msg string) {
			defer wg.Done()
			defer logger.Log(msg)
			var sclient *mongo.Client
			if sclient, err = NewMongoClient(uri, connString.SSLCaFile, connString.SSLClientCertificateKeyFile); err != nil {
				logger.Log(err.Error())
				return
			}
			defer sclient.Disconnect(context.Background())
			stats := NewStats(p.version)
			var server ClusterDetails
			if server, err = stats.GetClusterStatsSummary(sclient); err != nil {
				logger.Log(err.Error())
				return
			}
			mu.Lock()
			node := smap[server.ServerStatus.Repl.SetName]
			node.Servers = append(node.Servers, server)
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
