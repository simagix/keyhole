// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/simagix/gox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoCluster holds mongo cluster info
type MongoCluster struct {
	client   *mongo.Client
	cluster  bson.M
	verbose  bool
	filename string
}

// NewMongoCluster server info struct
func NewMongoCluster(client *mongo.Client) *MongoCluster {
	return &MongoCluster{client: client}
}

// SetVerbose -
func (mc *MongoCluster) SetVerbose(verbose bool) {
	mc.verbose = verbose
}

// SetOutputFilename sets output file name
func (mc *MongoCluster) SetOutputFilename(filename string) {
	mc.filename = strings.Replace(filename, ":", "_", -1)
}

// GetClusterInfo -
func (mc *MongoCluster) GetClusterInfo() (bson.M, error) {
	var err error
	var cur *mongo.Cursor
	var ctx = context.Background()
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
	mc.cluster["process"] = info.Process
	mc.cluster["sharding"] = info.Sharding
	mc.cluster["storage"] = info.StorageSize
	mc.cluster["version"] = info.Version
	// hostInfo
	var hostInfo bson.M
	if hostInfo, err = RunAdminCommand(mc.client, "hostInfo"); err == nil {
		config["hostInfo"] = trimMap(hostInfo)
	} else {
		config["hostInfo"] = bson.M{"ok": 0, "error": err.Error()}
	}

	// getCmdLineOpts
	var getCmdLineOpts bson.M
	if getCmdLineOpts, err = RunAdminCommand(mc.client, "getCmdLineOpts"); err == nil {
		config["getCmdLineOpts"] = trimMap(getCmdLineOpts)
	} else {
		config["getCmdLineOpts"] = bson.M{"ok": 0, "error": err.Error()}
	}

	// buildInfo
	var buildInfo bson.M
	if buildInfo, err = RunAdminCommand(mc.client, "buildInfo"); err == nil {
		config["buildInfo"] = trimMap(buildInfo)
	} else {
		config["buildInfo"] = bson.M{"ok": 0, "error": err.Error()}
	}

	// ServerStatus
	var serverStatus bson.M
	if serverStatus, err = RunAdminCommand(mc.client, "serverStatus"); err == nil {
		config["serverStatus"] = trimMap(serverStatus)
	} else {
		config["serverStatus"] = bson.M{"ok": 0, "error": err.Error()}
	}

	// replSetGetStatus
	if info.Cluster == "replica" {
		var replSetGetStatus bson.M
		if replSetGetStatus, err = RunAdminCommand(mc.client, "replSetGetStatus"); err == nil {
			config["replSetGetStatus"] = trimMap(replSetGetStatus)
		} else {
			config["replSetGetStatus"] = bson.M{"ok": 0, "error": err.Error()}
		}
	}

	// usersInfo
	var usersInfo bson.M
	if usersInfo, err = RunAdminCommand(mc.client, "usersInfo"); err == nil {
		config["usersInfo"] = trimMap(usersInfo)
	} else {
		config["usersInfo"] = bson.M{"ok": 0, "error": err.Error()}
	}

	// rolesInfo
	var rolesInfo bson.M
	if rolesInfo, err = RunAdminCommand(mc.client, "rolesInfo"); err == nil {
		config["rolesInfo"] = trimMap(rolesInfo)
	} else {
		config["rolesInfo"] = bson.M{"ok": 0, "error": err.Error()}
	}

	// collections firstDoc (findOne), indexes, and stats
	dbNames, _ := ListDatabaseNames(mc.client)
	var databases = []bson.M{}
	for _, dbName := range dbNames {
		if dbName == "admin" || dbName == "config" || dbName == "local" {
			continue
		}
		if cur, err = mc.client.Database(dbName).ListCollections(ctx, bson.M{}); err != nil {
			return mc.cluster, err
		}
		defer cur.Close(ctx)
		var collections = []bson.M{}
		ir := NewIndexesReader(mc.client)

		for cur.Next(ctx) {
			var elem = bson.M{}
			if err = cur.Decode(&elem); err != nil {
				continue
			}
			collectionName := fmt.Sprintf("%v", elem["name"])
			ns := dbName + "." + collectionName
			collection := mc.client.Database(dbName).Collection(collectionName)

			// firstDoc, FindOne
			var firstDoc bson.M
			if err = collection.FindOne(ctx, bson.D{{}}).Decode(&firstDoc); err != nil {
				err = nil
				continue
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
			collections = append(collections, bson.M{"NS": ns, "collection": collectionName, "document": firstDoc,
				"indexes": indexes, "stats": trimMap(stats)})
		}
		var stats bson.M
		stats, _ = RunCommandOnDB(mc.client, "dbStats", dbName)
		databases = append(databases, bson.M{"DB": dbName, "collections": collections, "stats": trimMap(stats)})
	}
	mc.cluster["databases"] = databases
	if err = gox.OutputGzipped([]byte(gox.Stringify(mc.cluster)), mc.filename); err == nil {
		fmt.Println("JSON is written to", mc.filename)
	}
	return mc.cluster, err
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
