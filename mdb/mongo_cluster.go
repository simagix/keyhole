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
	mc.cluster = bson.M{}
	config := GetServerInfo(mc.client)
	summary := config["summary"].(bson.M)
	mc.cluster["cluster"] = summary["cluster"]
	mc.cluster["host"] = summary["host"]
	mc.cluster["process"] = summary["process"]
	mc.cluster["version"] = summary["version"]
	mc.cluster["summary"] = getClusterSummaryString(summary)
	if mc.verbose {
		log.Println(mc.cluster["summary"])
	} else {
		return mc.cluster, err
	}
	delete(config, "summary")
	mc.cluster["config"] = config
	mc.SetFilename(fmt.Sprintf(`%v-cluster.bson.gz`, mc.cluster["host"]))
	clusterType := fmt.Sprintf(`%v`, mc.cluster["cluster"])
	if clusterType == SHARDED {
		var shards []ShardDoc
		if shards, err = GetShards(mc.client); err != nil {
			log.Println(err)
		}
		mc.cluster["shardIDs"] = shards
		var shardList []string
		if shardList, err = GetShardListWithURI(shards, mc.connString); err == nil {
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
					cluster := GetServerInfo(client)
					mu.Lock()
					shards = append(shards, cluster)
					mu.Unlock()
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
	if mc.verbose {
		log.Println("* collectServerInfo")
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
