// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// SHARDED cluster
const SHARDED = "sharded"

// REPLICA set
const REPLICA = "replica"

// STANDALONE cluster
const STANDALONE = "standalone"

// ServerInfo constains server info from db.serverStatus()
type ServerInfo struct {
	Cluster     string `json:"cluster" bson:"cluster"`
	Host        string `json:"host" bson:"host"`
	Process     string `json:"process" bson:"process"`
	Version     string `json:"version" bson:"version"`
	Sharding    bson.M `json:"sharding" bson:"sharding"`
	Repl        bson.M `json:"repl" bson:"repl"`
	StorageSize bson.M
}

// GetServerInfo returns ServerInfo from db.serverStatus()
func GetServerInfo(client *mongo.Client) (ServerInfo, error) {
	var err error
	var result bson.M
	var serverInfo = ServerInfo{}
	if result, err = RunAdminCommand(client, "serverStatus"); err != nil {
		return serverInfo, err
	}
	b, _ := bson.Marshal(result)
	bson.Unmarshal(b, &serverInfo)

	if serverInfo.Sharding == nil {
		serverInfo.Sharding = bson.M{}
	}

	if serverInfo.Process == "mongos" {
		serverInfo.Cluster = SHARDED
	} else if serverInfo.Repl != nil {
		serverInfo.Cluster = REPLICA
		serverInfo.Repl["oplog"] = GetOplogStats(client)
	} else {
		serverInfo.Cluster = STANDALONE
		serverInfo.Repl = bson.M{}
	}
	var names []string
	if names, err = ListDatabaseNames(client); err != nil {
		return serverInfo, err
	}

	dbStats := DBStats{}
	var dataSize, indexSize int
	list := []bson.M{}
	total := len(names)
	for i, name := range names {
		fmt.Fprintf(os.Stderr, "\r%3d%% ", (100*i)/total)
		result, _ = RunCommandOnDB(client, "dbStats", name)
		b, _ := json.Marshal(result)
		json.Unmarshal(b, &dbStats)
		dataSize += dbStats.DataSize
		indexSize += dbStats.IndexSize
		list = append(list, bson.M{"db": name, "objects": dbStats.Objects, "dataSize": dbStats.DataSize, "indexSize": dbStats.IndexSize})
	}
	fmt.Fprintf(os.Stderr, "\r     \r")
	serverInfo.StorageSize = bson.M{"totalDataSize (MB)": dataSize / 1024 / 1024, "totalIndexSize (MB)": indexSize / 1024 / 1024, "statsDetails": list}
	return serverInfo, nil
}

// ListDatabaseNames gets all database names
func ListDatabaseNames(client *mongo.Client) ([]string, error) {
	var err error
	var names []string
	var result mongo.ListDatabasesResult
	if result, err = client.ListDatabases(context.Background(), bson.M{}); err != nil {
		return names, err
	}
	for _, db := range result.Databases {
		names = append(names, db.Name)
	}
	return names, err
}
