// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"fmt"
	"log"
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
	bsonExt     string
	htmlExt     string
	keyholeInfo *KeyholeInfo
	redact      bool
	verbose     bool
}

// ClusterDetails stores cluster details
type ClusterDetails struct {
	BuildInfo        BuildInfo        `bson:"buildInfo"`
	CmdLineOpts      CmdLineOpts      `json:"getCmdLineOpts" bson:"getCmdLineOpts"`
	Cluster          string           `bson:"cluster"`
	Databases        []Database       `bson:"databases"`
	Host             string           `bson:"host"`
	HostInfo         HostInfo         `bson:"hostInfo"`
	KeyholeInfo      *KeyholeInfo     `bson:"keyhole"`
	OplogStats       OplogStats       `bson:"oplog"`
	Process          string           `bson:"process"`
	ReplSetGetStatus ReplSetGetStatus `bson:"replSetGetStatus"`
	ServerStatus     ServerStatus     `bson:"serverStatus"`
	Shards           []Shard          `bson:"shards"`
	Version          string           `bson:"version"`
}

// NewStats -
func NewStats(version string) *ClusterStats {
	s := ClusterStats{keyholeInfo: NewKeyholeInfo(version, "-allinfo"), bsonExt: "-stats.bson.gz", htmlExt: "-stats.html"}
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
	var cluster = ClusterDetails{}
	p.keyholeInfo.Log("GetClusterStats() begins")
	if cluster, err = p.GetClusterStatsSummary(client); err != nil {
		p.keyholeInfo.Log(fmt.Sprintf(`GetClusterStatsSummary(): %v`, err))
	}
	if cluster.CmdLineOpts, err = GetCmdLineOpts(client); err != nil {
		p.keyholeInfo.Log(fmt.Sprintf(`GetCmdLineOpts(): %v`, err))
	}
	if cluster.Cluster == Sharded { //collects from the primary of each shard
		message := "sharded detected, collecting from all servers"
		p.keyholeInfo.Log(message)
		if cluster.Shards, err = GetShards(client); err != nil {
			p.keyholeInfo.Log(fmt.Sprintf(`GetShards(): %v`, err))
		}
		if cluster.Shards, err = p.GeterversStatsSummary(cluster.Shards, connString); err != nil {
			p.keyholeInfo.Log(fmt.Sprintf(`GeterversStatsSummary(): %v`, err))
		}
	} else if cluster.Cluster == Replica && cluster.Process == "mongod" { //collects replica info
		message := "replica detected, collecting from all servers"
		p.keyholeInfo.Log(message)
		if cluster.ReplSetGetStatus, err = GetReplSetGetStatus(client); err != nil {
			p.keyholeInfo.Log(fmt.Sprintf(`GetReplSetGetStatus(): %v`, err))
		}

		setName := cluster.ServerStatus.Repl.SetName
		s := fmt.Sprintf(`%v/%v`, setName, strings.Join(cluster.ServerStatus.Repl.Hosts, ","))
		oneShard := []Shard{Shard{ID: setName, State: 1, Host: s}}
		if cluster.Shards, err = p.GeterversStatsSummary(oneShard, connString); err != nil {
			p.keyholeInfo.Log(fmt.Sprintf(`GeterversStatsSummary(): %v`, err))
		}
	}
	db := NewDatabaseStats()
	var databases []Database
	if databases, err = db.GetAllDatabasesStats(client); err != nil {
		p.keyholeInfo.Log(fmt.Sprintf(`GetAllDatabasesStats(): %v`, err))
	}
	for _, m := range db.GetLogs() {
		p.keyholeInfo.Add(m)
	}
	cluster.Databases = databases
	cluster.KeyholeInfo = p.keyholeInfo
	return cluster, nil
}

// Save saves to data
func (p *ClusterStats) Save(cluster ClusterDetails) (string, error) {
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
	var cluster = ClusterDetails{}
	if cluster.BuildInfo, err = GetBuildInfo(client); err != nil {
		p.keyholeInfo.Log(fmt.Sprintf(`GetBuildInfo(): %v`, err))
	}
	cluster.Version = cluster.BuildInfo.Version
	if cluster.HostInfo, err = GetHostInfo(client); err != nil {
		p.keyholeInfo.Log(fmt.Sprintf(`GetHostInfo(): %v`, err))
	}
	if cluster.ServerStatus, err = GetServerStatus(client); err != nil {
		p.keyholeInfo.Log(fmt.Sprintf(`GetServerStatus(): %v`, err))
	}
	cluster.Host = cluster.ServerStatus.Host
	cluster.Process = cluster.ServerStatus.Process
	cluster.Cluster = GetClusterType(cluster.ServerStatus)
	if cluster.Cluster == Replica && cluster.Process == "mongod" { //collects replica info
		if cluster.OplogStats, err = GetOplogStats(client); err != nil {
			p.keyholeInfo.Log(fmt.Sprintf(`GetOplogStats(): %v`, err))
		}
	}
	return cluster, nil
}

// GetClusterShortSummary returns one line summary
func (p *ClusterStats) GetClusterShortSummary(client *mongo.Client) string {
	var err error
	var c ClusterDetails
	if c, err = p.GetClusterStatsSummary(client); err != nil {
		p.keyholeInfo.Log(fmt.Sprintf(`GetClusterStatsSummary(): %v`, err))
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

// GeterversStatsSummary returns cluster stats from all shards
func (p *ClusterStats) GeterversStatsSummary(shards []Shard, connString connstring.ConnString) ([]Shard, error) {
	var err error
	var uris []string
	var smap = map[string]Shard{}
	for _, v := range shards {
		v.Servers = []ClusterDetails{}
		smap[v.ID] = v
	}
	if uris, err = GetAllServerURIs(shards, connString); err != nil {
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
		p.keyholeInfo.Log(msg)
		msg = fmt.Sprintf(`[t-%d] end collecting from %v`, i, s)
		wg.Add(1)
		go func(uri string, msg string) {
			defer wg.Done()
			defer p.keyholeInfo.Log(msg)
			var sclient *mongo.Client
			if sclient, err = NewMongoClient(uri, connString.SSLCaFile, connString.SSLClientCertificateKeyFile); err != nil {
				log.Println(err)
				return
			}
			defer sclient.Disconnect(context.Background())
			stats := NewStats(p.keyholeInfo.Version)
			var server ClusterDetails
			if server, err = stats.GetClusterStatsSummary(sclient); err != nil {
				log.Println(err)
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
