// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"sort"
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
	logfile    string
	verbose    bool
}

const replica = "replica"

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
	log.Println("traces are written to", mc.logfile)
	defer f.Close()
	log.SetOutput(f)
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
	mc.SetFilename(info.Host + ".bson.gz")
	mc.cluster["process"] = info.Process
	if info.Cluster == SHARDED {
		mc.cluster["sharding"] = info.Sharding
		if mc.cluster["shardIDs"], err = GetShards(mc.client); err != nil {
			log.Println("ERROR", err, "from buildInfo")
		}
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
					if sinfo.Cluster == replica {
						cluster["oplog"] = sinfo.Repl["oplog"]
					}
					if err = collectServerInfo(client, &cluster, sinfo.Cluster); err != nil {
						log.Println("ERROR", err, "from collectServerInfo")
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
		log.Println("ERROR", err, "from collectServerInfo")
	}

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
				continue
			}
			firstDoc = emptyBinData(firstDoc)
			indexes := ir.GetIndexesFromCollection(collection)

			// stats
			var stats bson.M
			mc.client.Database(dbName).RunCommand(ctx, bson.D{{Key: "collStats", Value: collectionName}}).Decode(&stats)
			if stats["shards"] != nil {
				for k := range stats["shards"].(primitive.M) {
					m := (stats["shards"].(primitive.M)[k]).(primitive.M)
					delete(m, "$clusterTime")
					delete(m, "$gleStats")
				}
			}
			log.Println(gox.Stringify(stats, "", "  "))
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
	log.SetOutput(os.Stdin)
	var data []byte
	if data, err = bson.Marshal(mc.cluster); err == nil {
		if err = gox.OutputGzipped(data, mc.filename); err == nil {
			fmt.Println("BSON is written to", mc.filename)
		}
	}
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
		log.Println("ERROR", err, "from hostInfo")
	}
	log.Println("hostInfo:\n", gox.Stringify((*cluster)["hostInfo"], "", "  "))
	// getCmdLineOpts
	var getCmdLineOpts bson.M
	if getCmdLineOpts, err = RunAdminCommand(client, "getCmdLineOpts"); err == nil {
		(*cluster)["getCmdLineOpts"] = trimMap(getCmdLineOpts)
	} else {
		(*cluster)["getCmdLineOpts"] = bson.M{"ok": 0, "error": err.Error()}
		log.Println("ERROR", err, "from getCmdLineOpts")
	}
	log.Println("getCmdLineOpts:\n", gox.Stringify((*cluster)["getCmdLineOpts"], "", "  "))
	// buildInfo
	var buildInfo bson.M
	if buildInfo, err = RunAdminCommand(client, "buildInfo"); err == nil {
		(*cluster)["buildInfo"] = trimMap(buildInfo)
	} else {
		(*cluster)["buildInfo"] = bson.M{"ok": 0, "error": err.Error()}
		log.Println("ERROR", err, "from buildInfo")
	}
	log.Println("buildInfo:\n", gox.Stringify((*cluster)["buildInfo"], "", "  "))
	// ServerStatus
	var serverStatus bson.M
	if serverStatus, err = RunAdminCommand(client, "serverStatus"); err == nil {
		(*cluster)["serverStatus"] = trimMap(serverStatus)
	} else {
		(*cluster)["serverStatus"] = bson.M{"ok": 0, "error": err.Error()}
		log.Println("ERROR", err, "from serverStatus")
	}
	log.Println("serverStatus:\n", gox.Stringify((*cluster)["serverStatus"], "", "  "))
	// replSetGetStatus
	if clusterType == replica {
		var replSetGetStatus bson.M
		if replSetGetStatus, err = RunAdminCommand(client, "replSetGetStatus"); err == nil {
			(*cluster)["replSetGetStatus"] = trimMap(replSetGetStatus)
		} else {
			(*cluster)["replSetGetStatus"] = bson.M{"ok": 0, "error": err.Error()}
			log.Println("ERROR", err, "from replSetGetStatus")
		}
	}
	log.Println("replSetGetStatus:\n", gox.Stringify((*cluster)["replSetGetStatus"], "", "  "))
	// usersInfo
	var usersInfo bson.M
	if usersInfo, err = RunAdminCommand(client, "usersInfo"); err == nil {
		(*cluster)["usersInfo"] = trimMap(usersInfo)
	} else {
		(*cluster)["usersInfo"] = bson.M{"ok": 0, "error": err.Error()}
		log.Println("ERROR", err, "from usersInfo")
	}
	log.Println("usersInfo:\n", gox.Stringify((*cluster)["usersInfo"], "", "  "))
	// rolesInfo
	var rolesInfo bson.M
	if rolesInfo, err = RunAdminCommand(client, "rolesInfo"); err == nil {
		(*cluster)["rolesInfo"] = trimMap(rolesInfo)
	} else {
		(*cluster)["rolesInfo"] = bson.M{"ok": 0, "error": err.Error()}
		log.Println("ERROR", err, "from rolesInfo")
	}
	log.Println("rolesInfo:\n", gox.Stringify((*cluster)["rolesInfo"], "", "  "))
	return err
}

func emptyBinData(firstDoc bson.M) bson.M {
	for k, v := range firstDoc {
		if reflect.TypeOf(v) == nil {
			continue
		}
		t := reflect.TypeOf(v).String()
		if t == "primitive.Binary" {
			firstDoc[k] = primitive.Binary{}
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
