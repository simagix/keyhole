// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// SHARDED cluster
const SHARDED = "sharded"

// REPLICA set
const REPLICA = "replica"

// STANDALONE cluster
const STANDALONE = "standalone"

// GetClusterSummary returns cluster summary
func GetClusterSummary(client *mongo.Client) string {
	doc := GetServerInfo(client)
	summary := doc["summary"].(bson.M)
	return getClusterSummaryString(summary)
}

func getClusterSummaryString(doc bson.M) string {
	return fmt.Sprintf(`MongoDB v%v %v %v (%v) %v %v %v cores %v mem`,
		doc["version"], doc["edition"], doc["host"], doc["os-name"],
		doc["process"], doc["cluster"], doc["numCores"], doc["memSizeMB"])
}

// GetServerInfo gets mongo server information
func GetServerInfo(client *mongo.Client) bson.M {
	var err error // hostInfo
	var estrs []string
	var cluster = bson.M{}
	var hostInfo bson.M
	if hostInfo, err = RunAdminCommand(client, "hostInfo"); err == nil {
		cluster["hostInfo"] = trimMap(hostInfo)
	} else {
		log.Println(err)
		estrs = append(estrs, err.Error())
		cluster["hostInfo"] = bson.M{"ok": 0, "error": err.Error()}
	}
	// getCmdLineOpts
	var getCmdLineOpts bson.M
	if getCmdLineOpts, err = RunAdminCommand(client, "getCmdLineOpts"); err == nil {
		cluster["getCmdLineOpts"] = trimMap(getCmdLineOpts)
	} else {
		log.Println(err)
		estrs = append(estrs, err.Error())
		cluster["getCmdLineOpts"] = bson.M{"ok": 0, "error": err.Error()}
	}
	// buildInfo
	var buildInfo bson.M
	if buildInfo, err = RunAdminCommand(client, "buildInfo"); err == nil {
		cluster["buildInfo"] = trimMap(buildInfo)
	} else {
		log.Println(err)
		estrs = append(estrs, err.Error())
		cluster["buildInfo"] = bson.M{"ok": 0, "error": err.Error()}
	}
	// ServerStatus
	var serverStatus bson.M
	if serverStatus, err = RunAdminCommand(client, "serverStatus"); err == nil {
		cluster["serverStatus"] = trimMap(serverStatus)
	} else {
		log.Println(err)
		estrs = append(estrs, err.Error())
		cluster["serverStatus"] = bson.M{"ok": 0, "error": err.Error()}
	}

	var bi struct {
		Version string
		Modules []string
	}
	var hi struct {
		OS struct {
			Name string
		}
		System struct {
			Hostname  string
			NumCores  int
			MemSizeMB int
		}
	}
	var ss struct {
		Process string
		Repl    struct {
			SetName string
		}
		Sharding struct {
			ConfigsvrConnectionString string
		}
	}
	var data []byte
	data, _ = bson.Marshal(buildInfo)
	bson.Unmarshal(data, &bi)
	edition := "community"
	if len(bi.Modules) > 0 {
		edition = bi.Modules[0]
	}
	data, _ = bson.Marshal(hostInfo)
	bson.Unmarshal(data, &hi)
	data, _ = bson.Marshal(serverStatus)
	bson.Unmarshal(data, &ss)
	clusterType := "standalone"
	if ss.Repl.SetName != "" {
		clusterType = "replica"
	} else if ss.Sharding.ConfigsvrConnectionString != "" {
		clusterType = "sharded"
	}
	// replSetGetStatus
	if clusterType == replica {
		var replSetGetStatus bson.M
		if replSetGetStatus, err = RunAdminCommand(client, "replSetGetStatus"); err == nil {
			cluster["replSetGetStatus"] = trimMap(replSetGetStatus)
		} else {
			log.Println(err)
			estrs = append(estrs, err.Error())
			cluster["replSetGetStatus"] = bson.M{"ok": 0, "error": err.Error()}
		}
		cluster["oplog"] = GetOplogStats(client)
	}
	summary := bson.M{}
	summary["cluster"] = clusterType
	summary["host"] = hi.System.Hostname
	summary["process"] = ss.Process
	summary["edition"] = edition
	summary["version"] = bi.Version
	summary["os-name"] = hi.OS.Name
	summary["numCores"] = hi.System.NumCores
	summary["memSizeMB"] = hi.System.MemSizeMB
	cluster["summary"] = summary
	cluster["cluster"] = clusterType
	return cluster
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
