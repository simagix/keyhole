// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/simagix/gox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

// MongoCluster holds mongo cluster info
type MongoCluster struct {
	client     *mongo.Client
	cluster    bson.M
	connString connstring.ConnString
	doodle     bool
	filename   string
	verbose    bool
}

const replica = "replica"

// NewMongoCluster server info struct
func NewMongoCluster(client *mongo.Client) *MongoCluster {
	hostname, _ := os.Hostname()
	return &MongoCluster{client: client, filename: hostname + ".bson.gz"}
}

// SetVerbose -
func (mc *MongoCluster) SetVerbose(verbose bool) {
	mc.verbose = verbose
}

// SetFilename sets output file name
func (mc *MongoCluster) SetFilename(filename string) {
	mc.filename = strings.Replace(filename, ":", "_", -1)
}

// SetConnString set connString object
func (mc *MongoCluster) SetConnString(connString connstring.ConnString) {
	mc.connString = connString
}

// SetDoodleMode sets doodle
func (mc *MongoCluster) SetDoodleMode(doodle bool) {
	mc.doodle = doodle
}

// GetClusterInfo -
func (mc *MongoCluster) GetClusterInfo() (bson.M, error) {
	var err error
	var config = bson.M{}
	mc.cluster = bson.M{"config": config}
	var info ServerInfo
	if info, err = GetServerInfo(mc.client); err != nil {
		return nil, err
	}
	var val bson.M
	b, _ := bson.Marshal(info)
	bson.Unmarshal(b, &val)
	if mc.verbose == false {
		return val, err
	}
	info.StorageSize["databases"] = info.StorageSize["statsDetails"]
	delete(info.StorageSize, "statsDetails")
	mc.cluster["cluster"] = info.Cluster
	mc.cluster["host"] = info.Host
	mc.SetFilename(info.Host + ".bson.gz")
	mc.cluster["process"] = info.Process
	if info.Cluster == SHARDED {
		mc.cluster["sharding"] = info.Sharding
		if mc.cluster["shardIDs"], err = GetShards(mc.client); err != nil {
		}
		var shardList []string
		if shardList, err = GetShardListWithURI(mc.client, mc.connString.String()); err == nil {
			var shards []bson.M
			for _, shardURI := range shardList {
				var client *mongo.Client
				if client, err = NewMongoClient(shardURI, mc.connString.SSLCaFile, mc.connString.SSLClientCertificateKeyFile); err != nil {
					continue
				}
				var sinfo ServerInfo
				if sinfo, err = GetServerInfo(client, true); err == nil {
					cluster := bson.M{}
					cluster["cluster"] = sinfo.Cluster
					cluster["host"] = sinfo.Host
					cluster["process"] = sinfo.Process
					if sinfo.Cluster == replica {
						cluster["oplog"] = sinfo.Repl["oplog"]
					}
					if err = collectServerInfo(client, &cluster, sinfo.Cluster); err != nil {
						continue
					}
					shards = append(shards, cluster)
				}
			}
			mc.cluster["shards"] = shards
		}
	}
	mc.cluster["storage"] = info.StorageSize
	mc.cluster["version"] = info.Version
	if info.Cluster == replica {
		config["oplog"] = info.Repl["oplog"]
	}
	if err = collectServerInfo(mc.client, &config, info.Cluster); err != nil {
		return mc.cluster, err
	}
	fmt.Fprintf(os.Stderr, "\r     \r")
	if mc.cluster["databases"], err = GetAllDatabasesInfo(mc.client); err != nil {
		return mc.cluster, err
	}
	var data []byte
	if data, err = bson.Marshal(mc.cluster); err != nil {
		return mc.cluster, err
	} else if err = gox.OutputGzipped(data, mc.filename); err != nil {
		return mc.cluster, err
	}
	fmt.Println("\rBSON is written to", mc.filename)
	return mc.cluster, err
}

func collectServerInfo(client *mongo.Client, cluster *bson.M, clusterType string) error {
	var err error // hostInfo
	var hostInfo bson.M
	if *cluster == nil {
		cluster = &bson.M{}
	}
	if hostInfo, err = RunAdminCommand(client, "hostInfo"); err == nil {
		(*cluster)["hostInfo"] = trimMap(hostInfo)
	} else {
		(*cluster)["hostInfo"] = bson.M{"ok": 0, "error": err.Error()}
	}
	// getCmdLineOpts
	var getCmdLineOpts bson.M
	if getCmdLineOpts, err = RunAdminCommand(client, "getCmdLineOpts"); err == nil {
		(*cluster)["getCmdLineOpts"] = trimMap(getCmdLineOpts)
	} else {
		(*cluster)["getCmdLineOpts"] = bson.M{"ok": 0, "error": err.Error()}
	}
	// buildInfo
	var buildInfo bson.M
	if buildInfo, err = RunAdminCommand(client, "buildInfo"); err == nil {
		(*cluster)["buildInfo"] = trimMap(buildInfo)
	} else {
		(*cluster)["buildInfo"] = bson.M{"ok": 0, "error": err.Error()}
	}
	// ServerStatus
	var serverStatus bson.M
	if serverStatus, err = RunAdminCommand(client, "serverStatus"); err == nil {
		(*cluster)["serverStatus"] = trimMap(serverStatus)
	} else {
		(*cluster)["serverStatus"] = bson.M{"ok": 0, "error": err.Error()}
	}
	// replSetGetStatus
	if clusterType == replica {
		var replSetGetStatus bson.M
		if replSetGetStatus, err = RunAdminCommand(client, "replSetGetStatus"); err == nil {
			(*cluster)["replSetGetStatus"] = trimMap(replSetGetStatus)
		} else {
			(*cluster)["replSetGetStatus"] = bson.M{"ok": 0, "error": err.Error()}
		}
	}
	// usersInfo
	var usersInfo bson.M
	if usersInfo, err = RunAdminCommand(client, "usersInfo"); err == nil {
		(*cluster)["usersInfo"] = trimMap(usersInfo)
	} else {
		(*cluster)["usersInfo"] = bson.M{"ok": 0, "error": err.Error()}
	}
	// rolesInfo
	var rolesInfo bson.M
	if rolesInfo, err = RunAdminCommand(client, "rolesInfo"); err == nil {
		(*cluster)["rolesInfo"] = trimMap(rolesInfo)
	} else {
		(*cluster)["rolesInfo"] = bson.M{"ok": 0, "error": err.Error()}
	}
	return err
}

func emptyBinData(firstDoc bson.M) bson.M {
	for k, v := range firstDoc {
		if reflect.TypeOf(v) == nil {
			continue
		}
		t := reflect.TypeOf(v).String()
		if t == "primitive.Binary" {
			if v.(primitive.Binary).Subtype != 4 { // empty data when it's not UUID
				firstDoc[k] = primitive.Binary{Subtype: v.(primitive.Binary).Subtype}
			}
		} else {
			// fmt.Println(v, t)
		}
	}
	return firstDoc
}

func trimMap(doc bson.M) bson.M {
	delete(doc, "$clusterTime")
	delete(doc, "operationTime")
	delete(doc, "$gleStats")
	delete(doc, "ok")
	return doc
}

func round(s string) string {
	idx := strings.Index(s, ".")
	if idx < 0 {
		return s
	}
	return s[:idx+2]
}
