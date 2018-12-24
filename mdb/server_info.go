// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"encoding/json"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
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
	var stat = ServerStatusDoc{}
	if result, err = RunAdminCommand(client, "serverStatus"); err != nil {
		return serverInfo, err
	}
	b, _ := bson.Marshal(result)
	bson.Unmarshal(b, &stat)
	serverInfo.Host = stat.Host
	serverInfo.Process = stat.Process
	serverInfo.Version = stat.Version
	serverInfo.Sharding = bson.M{}
	if stat.Sharding != nil {
		serverInfo.Sharding = stat.Sharding
	}
	serverInfo.Repl = bson.M{}
	if stat.Repl != nil {
		serverInfo.Repl = stat.Repl
	}

	if stat.Process == "mongos" {
		serverInfo.Cluster = SHARDED
	} else if stat.Repl != nil {
		serverInfo.Cluster = REPLICA
	} else {
		serverInfo.Cluster = STANDALONE
	}

	var names []string
	if names, err = ListDatabaseNames(client); err != nil {
		return serverInfo, err
	}

	dbStats := DBStats{}
	var dataSize, indexSize int
	list := []bson.M{}

	for _, name := range names {
		result, _ = RunCommandOnDB(client, "dbStats", name)
		b, _ := json.Marshal(result)
		json.Unmarshal(b, &dbStats)
		dataSize += dbStats.DataSize
		indexSize += dbStats.IndexSize
		list = append(list, bson.M{"db": name, "objects": dbStats.Objects, "dataSize": dbStats.DataSize, "indexSize": dbStats.IndexSize})
	}

	serverInfo.StorageSize = bson.M{"totalDataSize (MB)": dataSize / 1024 / 1024, "totalIndexSize (MB)": indexSize / 1024 / 1024, "statsDetails": list}
	return serverInfo, nil
}

// GetDatabaseNames gets all database names
// Doesn't work on Atlas b/c of filter priviledge
func GetDatabaseNames(client *mongo.Client) ([]string, error) {
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

// ListDatabaseNames gets all database names
func ListDatabaseNames(client *mongo.Client) ([]string, error) {
	var err error
	var m bson.M
	var names []string
	if m, err = RunAdminCommand(client, "listDatabases"); err != nil {
		return names, err
	}
	for _, db := range m["databases"].(primitive.A) {
		names = append(names, (db.(bson.M))["name"].(string))
	}
	return names, err
}
