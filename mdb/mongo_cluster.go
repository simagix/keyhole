// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/simagix/gox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

// MongoCluster holds mongo cluster info
type MongoCluster struct {
	client      *mongo.Client
	cluster     bson.M
	conns       int
	connString  connstring.ConnString
	filename    string
	KeyholeInfo *KeyholeInfo
	redaction   bool
	verbose     bool
	vv          bool
}

const replica = "replica"

// NewMongoCluster server info struct
func NewMongoCluster(client *mongo.Client) *MongoCluster {
	hostname, _ := os.Hostname()
	return &MongoCluster{client: client, filename: hostname + "-cluster.bson.gz",
		KeyholeInfo: &KeyholeInfo{}}
}

// SetKeyholeInfo sets keyhole version
func (mc *MongoCluster) SetKeyholeInfo(keyholeInfo *KeyholeInfo) {
	mc.KeyholeInfo = keyholeInfo
}

// SetRedaction sets redact
func (mc *MongoCluster) SetRedaction(redaction bool) {
	mc.redaction = redaction
}

// SetVerbose -
func (mc *MongoCluster) SetVerbose(verbose bool) {
	mc.verbose = verbose
}

// SetVeryVerbose -
func (mc *MongoCluster) SetVeryVerbose(vv bool) {
	mc.vv = vv
	if mc.vv == true {
		mc.verbose = true
	}
}

// SetFilename sets output file name
func (mc *MongoCluster) SetFilename(filename string) {
	mc.filename = strings.Replace(filename, ":", "_", -1)
}

// SetNumberConnections set # of conns
func (mc *MongoCluster) SetNumberConnections(conns int) {
	mc.conns = conns
}

// SetConnString set connString object
func (mc *MongoCluster) SetConnString(connString connstring.ConnString) {
	mc.connString = connString
}

// GetClusterInfo -
func (mc *MongoCluster) GetClusterInfo() (bson.M, error) {
	var err error
	var config = bson.M{}
	mc.cluster = bson.M{"config": config}
	var info ServerInfo
	if mc.verbose {
		log.Println("* GetClusterInfo")
	}
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
	mc.SetFilename(info.Host + "-cluster.bson.gz")
	mc.cluster["process"] = info.Process
	if info.Cluster == SHARDED {
		mc.cluster["sharding"] = info.Sharding
		if mc.cluster["shardIDs"], err = GetShards(mc.client); err != nil {
			log.Println(err)
		}
		var shardList []string
		if shardList, err = GetShardListWithURI(mc.client, mc.connString.String()); err == nil {
			var mu sync.Mutex
			var wg = gox.NewWaitGroup(mc.conns) // runs in parallel
			var shards []bson.M
			for i, shardURI := range shardList {
				wg.Add(1)
				go func(shardURI string, i int) {
					defer wg.Done()
					s := shardURI
					if mc.connString.Password != "" {
						s = strings.ReplaceAll(s, mc.connString.Password, "xxxxxx")
					}
					msg := fmt.Sprintf(`[t-%d] begin collecting from %v`, i, s)
					log.Println(msg)
					mu.Lock()
					mc.KeyholeInfo.Log(msg)
					mu.Unlock()
					var client *mongo.Client
					if client, err = NewMongoClient(shardURI, mc.connString.SSLCaFile, mc.connString.SSLClientCertificateKeyFile); err != nil {
						return
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
							return
						}
						mu.Lock()
						shards = append(shards, cluster)
						mu.Unlock()
					}
					msg = fmt.Sprintf(`[t-%d] end collecting from %v`, i, s)
					log.Println(msg)
					mu.Lock()
					mc.KeyholeInfo.Log(msg)
					mu.Unlock()
				}(shardURI, i)
			}
			wg.Wait()
			mc.cluster["shards"] = shards
		}
	}
	mc.cluster["storage"] = info.StorageSize
	mc.cluster["version"] = info.Version
	if info.Cluster == replica {
		config["oplog"] = info.Repl["oplog"]
	}
	if mc.verbose {
		log.Println("* collectServerInfo")
	}
	if err = collectServerInfo(mc.client, &config, info.Cluster); err != nil {
		return mc.cluster, err
	}
	dbi := NewDatabaseInfo()
	dbi.SetNumberConnections(mc.conns)
	dbi.SetRedaction(mc.redaction)
	dbi.SetVerbose(mc.verbose)
	dbi.SetVeryVerbose(mc.vv)
	if mc.cluster["databases"], err = dbi.GetAllDatabasesInfo(mc.client); err != nil {
		return mc.cluster, err
	}
	for _, s := range dbi.GetLogs() {
		mc.KeyholeInfo.Log(s)
	}
	mc.KeyholeInfo.Log("GetClusterInfo() ends")
	mc.cluster["keyhole"] = mc.KeyholeInfo
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
