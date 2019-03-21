// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoCluster holds mongo cluster info
type MongoCluster struct {
	client  *mongo.Client
	verbose bool
}

// NewMongoCluster server info struct
func NewMongoCluster(client *mongo.Client) *MongoCluster {
	return &MongoCluster{client: client}
}

// SetVerbose -
func (mc *MongoCluster) SetVerbose(verbose bool) {
	mc.verbose = verbose
}

// GetInfo -
func (mc *MongoCluster) GetInfo() (bson.M, error) {
	var err error
	var cur *mongo.Cursor
	var icur *mongo.Cursor
	var ctx = context.Background()
	var config = bson.M{}
	var cluster = bson.M{"config": config}
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
	cluster["cluster"] = info.Cluster
	cluster["host"] = info.Host
	cluster["process"] = info.Process
	cluster["sharding"] = info.Sharding
	cluster["storage"] = info.StorageSize
	cluster["version"] = info.Version
	// hostInfo
	var hostInfo bson.M
	if hostInfo, err = RunAdminCommand(mc.client, "hostInfo"); err != nil {
		return cluster, err
	}
	config["hostInfo"] = trimMap(hostInfo)

	// getCmdLineOpts
	var getCmdLineOpts bson.M
	if getCmdLineOpts, err = RunAdminCommand(mc.client, "getCmdLineOpts"); err != nil {
		return cluster, err
	}
	config["getCmdLineOpts"] = trimMap(getCmdLineOpts)

	// buildInfo
	var buildInfo bson.M
	if buildInfo, err = RunAdminCommand(mc.client, "buildInfo"); err != nil {
		return cluster, err
	}
	config["buildInfo"] = trimMap(buildInfo)

	// ServerStatus
	var serverStatus bson.M
	if serverStatus, err = RunAdminCommand(mc.client, "serverStatus"); err != nil {
		return cluster, err
	}
	config["serverStatus"] = trimMap(serverStatus)

	// replSetGetStatus
	if info.Cluster == "replica" {
		var replSetGetStatus bson.M
		if replSetGetStatus, err = RunAdminCommand(mc.client, "replSetGetStatus"); err != nil {
			return cluster, err
		}
		config["replSetGetStatus"] = trimMap(replSetGetStatus)
	}

	// usersInfo
	var usersInfo bson.M
	if usersInfo, err = RunAdminCommand(mc.client, "usersInfo"); err != nil {
		log.Println(err)
	}
	config["usersInfo"] = trimMap(usersInfo)

	// rolesInfo
	var rolesInfo bson.M
	if rolesInfo, err = RunAdminCommand(mc.client, "rolesInfo"); err != nil {
		log.Println(err)
	}
	config["rolesInfo"] = trimMap(rolesInfo)

	// collections firstDoc (findOne), indexes, and stats
	dbNames, err := ListDatabaseNames(mc.client)
	var databases = []bson.M{}
	for _, dbName := range dbNames {
		// if dbName == "admin" || dbName == "config" || dbName == "local" {
		// 	continue
		// }
		if cur, err = mc.client.Database(dbName).ListCollections(ctx, bson.M{}); err != nil {
			return cluster, err
		}
		defer cur.Close(ctx)
		var collections = []bson.M{}

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
				continue
			}

			// indexes
			view := collection.Indexes()
			if icur, err = view.List(ctx); err != nil {
				continue
			}

			indexes := []bson.M{}
			for icur.Next(ctx) {
				val := bson.M{}
				icur.Decode(&val)
				indexes = append(indexes, val)
			}
			icur.Close(ctx)

			// stats
			var stats bson.M
			err = mc.client.Database(dbName).RunCommand(ctx, bson.D{{Key: "collStats", Value: collectionName}}).Decode(&stats)
			delete(stats, "indexDetails")
			delete(stats, "wiredTiger")
			collections = append(collections, bson.M{"NS": ns, "collection": collectionName, "document": firstDoc, "indexes": indexes, "stats": trimMap(stats)})
		}
		var stats bson.M
		stats, _ = RunCommandOnDB(mc.client, "dbStats", dbName)
		databases = append(databases, bson.M{"DB": dbName, "collections": collections, "stats": trimMap(stats)})
	}
	cluster["databases"] = databases
	return cluster, err
}

func trimMap(doc bson.M) bson.M {
	delete(doc, "$clusterTime")
	delete(doc, "operationTime")
	delete(doc, "ok")
	return doc
}
