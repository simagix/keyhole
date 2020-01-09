// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/simagix/gox"
	"github.com/simagix/keyhole/sim/util"
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
	logfile    string
	verbose    bool
}

// NewMongoCluster server info struct
func NewMongoCluster(client *mongo.Client) *MongoCluster {
	hostname, _ := os.Hostname()
	return &MongoCluster{client: client, filename: hostname + ".json.gz", logfile: hostname + ".keyhole.log"}
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
	f, err := os.OpenFile(mc.logfile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
	log.Println("GetClusterInfo(), logs are written to", mc.logfile)
	var cur *mongo.Cursor
	var ctx = context.Background()
	var config = bson.M{}

	mc.cluster = bson.M{"config": config}
	var info ServerInfo
	log.Println("GetServerInfo()")
	if info, err = GetServerInfo(mc.client); err != nil {
		return nil, err
	}
	log.Println("server info:\n", gox.Stringify(info, "", "  "))
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
	mc.SetFilename(info.Host + ".json.gz")
	mc.cluster["process"] = info.Process
	if info.Cluster == SHARDED {
		mc.cluster["sharding"] = info.Sharding
		var shardList []string
		if shardList, err = GetShardListWithURI(mc.client, mc.connString.String()); err == nil {
			var shards []bson.M
			for _, shardURI := range shardList {
				var client *mongo.Client
				if client, err = NewMongoClient(shardURI, mc.connString.SSLCaFile, mc.connString.SSLClientCertificateKeyFile); err != nil {
					log.Println(err)
					continue
				}
				var sinfo ServerInfo
				if sinfo, err = GetServerInfo(client, true); err == nil {
					cluster := bson.M{}
					cluster["cluster"] = sinfo.Cluster
					cluster["host"] = sinfo.Host
					cluster["process"] = sinfo.Process
					if hostInfo, err := RunAdminCommand(mc.client, "hostInfo"); err == nil {
						cluster["hostInfo"] = trimMap(hostInfo)
					}
					if buildInfo, err := RunAdminCommand(mc.client, "buildInfo"); err == nil {
						cluster["buildInfo"] = trimMap(buildInfo)
					}
					if sinfo.Cluster == "replica" {
						cluster["oplog"] = sinfo.Repl["oplog"]
						var replSetGetStatus bson.M
						if replSetGetStatus, err = RunAdminCommand(mc.client, "replSetGetStatus"); err == nil {
							cluster["replSetGetStatus"] = trimMap(replSetGetStatus)
						}
					}
					shards = append(shards, cluster)
				}
			}
			mc.cluster["shards"] = shards
		}
	}
	mc.cluster["storage"] = info.StorageSize
	mc.cluster["version"] = info.Version

	// hostInfo
	var hostInfo bson.M
	if hostInfo, err = RunAdminCommand(mc.client, "hostInfo"); err == nil {
		config["hostInfo"] = trimMap(hostInfo)
	} else {
		config["hostInfo"] = bson.M{"ok": 0, "error": err.Error()}
		log.Println("ERROR", err, "from hostInfo")
	}
	log.Println("hostInfo:\n", gox.Stringify(config["hostInfo"], "", "  "))

	// getCmdLineOpts
	var getCmdLineOpts bson.M
	if getCmdLineOpts, err = RunAdminCommand(mc.client, "getCmdLineOpts"); err == nil {
		config["getCmdLineOpts"] = trimMap(getCmdLineOpts)
	} else {
		config["getCmdLineOpts"] = bson.M{"ok": 0, "error": err.Error()}
		log.Println("ERROR", err, "from getCmdLineOpts")
	}
	log.Println("getCmdLineOpts:\n", gox.Stringify(config["getCmdLineOpts"], "", "  "))

	// buildInfo
	var buildInfo bson.M
	if buildInfo, err = RunAdminCommand(mc.client, "buildInfo"); err == nil {
		config["buildInfo"] = trimMap(buildInfo)
	} else {
		config["buildInfo"] = bson.M{"ok": 0, "error": err.Error()}
		log.Println("ERROR", err, "from buildInfo")
	}
	log.Println("buildInfo:\n", gox.Stringify(config["buildInfo"], "", "  "))

	// ServerStatus
	var serverStatus bson.M
	if serverStatus, err = RunAdminCommand(mc.client, "serverStatus"); err == nil {
		config["serverStatus"] = trimMap(serverStatus)
	} else {
		config["serverStatus"] = bson.M{"ok": 0, "error": err.Error()}
		log.Println("ERROR", err, "from serverStatus")
	}
	log.Println("serverStatus:\n", gox.Stringify(config["serverStatus"], "", "  "))

	// replSetGetStatus
	if info.Cluster == "replica" {
		config["oplog"] = info.Repl["oplog"]
		var replSetGetStatus bson.M
		if replSetGetStatus, err = RunAdminCommand(mc.client, "replSetGetStatus"); err == nil {
			config["replSetGetStatus"] = trimMap(replSetGetStatus)
		} else {
			config["replSetGetStatus"] = bson.M{"ok": 0, "error": err.Error()}
			log.Println("ERROR", err, "from replSetGetStatus")
		}
	}
	log.Println("replSetGetStatus:\n", gox.Stringify(config["replSetGetStatus"], "", "  "))

	// usersInfo
	var usersInfo bson.M
	if usersInfo, err = RunAdminCommand(mc.client, "usersInfo"); err == nil {
		config["usersInfo"] = trimMap(usersInfo)
	} else {
		config["usersInfo"] = bson.M{"ok": 0, "error": err.Error()}
		log.Println("ERROR", err, "from usersInfo")
	}
	log.Println("usersInfo:\n", gox.Stringify(config["usersInfo"], "", "  "))

	// rolesInfo
	var rolesInfo bson.M
	if rolesInfo, err = RunAdminCommand(mc.client, "rolesInfo"); err == nil {
		config["rolesInfo"] = trimMap(rolesInfo)
	} else {
		config["rolesInfo"] = bson.M{"ok": 0, "error": err.Error()}
		log.Println("ERROR", err, "from rolesInfo")
	}
	log.Println("rolesInfo:\n", gox.Stringify(config["rolesInfo"], "", "  "))

	// collections firstDoc (findOne), indexes, and stats
	dbNames, err := ListDatabaseNames(mc.client)
	if err != nil {
		fmt.Println("ERROR", err, "from ListDatabaseNames")
		return nil, err
	}
	var databases = []bson.M{}
	total := len(dbNames)
	for i, dbName := range dbNames {
		fmt.Fprintf(os.Stderr, "\r%3d%% ", (100*i)/total)
		if dbName == "admin" || dbName == "config" || dbName == "local" {
			log.Println("skip", dbName)
			continue
		}
		if cur, err = mc.client.Database(dbName).ListCollections(ctx, bson.M{}); err != nil {
			log.Println("ERROR", err, "from ListCollections", dbName)
			return mc.cluster, err
		}
		defer cur.Close(ctx)
		var collections = []bson.M{}
		ir := NewIndexesReader(mc.client)
		collectionNames := []string{}

		for cur.Next(ctx) {
			var elem = bson.M{}
			if err = cur.Decode(&elem); err != nil {
				log.Println("ERROR", err, "from cur.Decode")
				continue
			}
			coll := fmt.Sprintf("%v", elem["name"])
			collType := fmt.Sprintf("%v", elem["type"])
			if strings.Index(coll, "system.") == 0 || (elem["type"] != nil && collType != "collection") {
				log.Println("ERROR from strings.Index, (collection, type):", coll, collType)
				continue
			}
			collectionNames = append(collectionNames, coll)
		}

		sort.Strings(collectionNames)
		for _, collectionName := range collectionNames {
			ns := dbName + "." + collectionName
			log.Println("ns:", ns)
			collection := mc.client.Database(dbName).Collection(collectionName)

			// firstDoc, FindOne
			var firstDoc bson.M
			if err = collection.FindOne(ctx, bson.D{{}}).Decode(&firstDoc); err != nil {
				log.Println("ERROR", err, "from", ns)
				err = nil
				continue
			}
			firstDoc = convertDecimal128ToFloa64(firstDoc)
			if mc.doodle == true {
				if cdoc, err := util.GetRandomizedDoc([]byte(gox.Stringify(firstDoc)), false); err == nil {
					firstDoc = cdoc
				} else {
					log.Println(err)
				}
			}

			// indexes
			indexes := ir.GetIndexesFromCollection(collection)

			// stats
			var stats bson.M
			mc.client.Database(dbName).RunCommand(ctx, bson.D{{Key: "collStats", Value: collectionName}}).Decode(&stats)
			delete(stats, "indexDetails")
			delete(stats, "wiredTiger")
			if stats["shards"] != nil {
				for k := range stats["shards"].(primitive.M) {
					m := (stats["shards"].(primitive.M)[k]).(primitive.M)
					delete(m, "$clusterTime")
					delete(m, "$gleStats")
					delete(m, "indexDetails")
					delete(m, "wiredTiger")
				}
			}
			log.Println(gox.Stringify(stats, "", "  "))
			if gox.Stringify(firstDoc) == "" {
				log.Println(firstDoc)
				firstDoc = bson.M{}
			}
			collections = append(collections, bson.M{"NS": ns, "collection": collectionName, "document": firstDoc,
				"indexes": indexes, "stats": trimMap(stats)})
			log.Println("collections processed", len(collections))
		}
		var stats bson.M
		if stats, err = RunCommandOnDB(mc.client, "dbStats", dbName); err != nil {
			fmt.Println("ERROR", err, "from RunCommandOnDB dbStats", dbName)
			continue
		}
		databases = append(databases, bson.M{"DB": dbName, "collections": collections, "stats": trimMap(stats)})
	}
	fmt.Fprintf(os.Stderr, "\r     \r")
	mc.cluster["databases"] = databases
	var data []byte
	if data, err = json.Marshal(mc.cluster); err != nil {
		log.Println(err)
		return mc.cluster, err
	}
	clusterJSON := string(data)
	log.Println("cluster info", len(clusterJSON))
	log.Println(clusterJSON)
	if err = gox.OutputGzipped([]byte(clusterJSON), mc.filename); err == nil {
		fmt.Println("JSON is written to", mc.filename)
	}
	return mc.cluster, err
}

func convertDecimal128ToFloa64(firstDoc bson.M) bson.M {
	for k, v := range firstDoc {
		if reflect.TypeOf(v) == nil {
			continue
		}
		t := reflect.TypeOf(v).String()
		if t == "primitive.Decimal128" {
			firstDoc[k], _ = strconv.ParseFloat(v.(primitive.Decimal128).String(), 64)
		} else if t == "primitive.M" {
			firstDoc[k] = convertDecimal128ToFloa64(v.(bson.M))
		} else if t == "primitive.Binary" {
			firstDoc[k] = primitive.Binary{}
			// } else if t == "string" && len(fmt.Sprintf("%v", v)) > 32 {
			// 	firstDoc[k] = fmt.Sprintf("string:%v", len(fmt.Sprintf("%v", v)))
		} else if t == "float64" && math.IsNaN(v.(float64)) {
			firstDoc[k] = float64(0)
		} else {
			// fmt.Println(v, t)
		}
	}
	return firstDoc
}

func trimMap(doc bson.M) bson.M {
	delete(doc, "$clusterTime")
	delete(doc, "operationTime")
	delete(doc, "ok")
	return doc
}

// GetStorageSize returns storage size in [TGMK] B
func GetStorageSize(num interface{}) string {
	f := fmt.Sprintf("%v", num)
	x, err := strconv.ParseFloat(f, 64)
	if err != nil {
		return f
	}

	if x >= (1024 * 1024 * 1024 * 1024) {
		s := fmt.Sprintf("%v", x/(1024*1024*1024*1024))
		return round(s) + " TB"
	} else if x >= (1024 * 1024 * 1024) {
		s := fmt.Sprintf("%v", x/(1024*1024*1024))
		return round(s) + " GB"
	} else if x >= (1024 * 1024) {
		s := fmt.Sprintf("%v", x/(1024*1024))
		return round(s) + " MB"
	} else if x >= 1024 {
		s := fmt.Sprintf("%v", x/1024)
		return round(s) + " KB"
	}
	s := fmt.Sprintf("%v", x)
	return round(s) + " B"
}

func round(s string) string {
	idx := strings.Index(s, ".")
	if idx < 0 {
		return s
	}
	return s[:idx+2]
}
