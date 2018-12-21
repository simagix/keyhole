// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mongo

import (
	"encoding/json"

	"github.com/globalsign/mgo"
	"github.com/mongodb/mongo-go-driver/bson"
)

// SHARDED cluster
const SHARDED = "sharded"

// REPLICA set
const REPLICA = "replica"

// STANDALONE cluster
const STANDALONE = "standalone"

// ServerInfo constains server info from db.serverStatus()
type ServerInfo struct {
	Cluster     string      `json:"cluster" bson:"cluster"`
	Host        string      `json:"host" bson:"host"`
	Process     string      `json:"process" bson:"process"`
	Version     string      `json:"version" bson:"version"`
	Sharding    interface{} `json:"sharding" bson:"sharding"`
	Repl        interface{} `json:"repl" bson:"repl"`
	StorageSize bson.M
}

// GetServerInfo returns ServerInfo from db.serverStatus()
func GetServerInfo(session *mgo.Session) (ServerInfo, error) {
	var err error
	var serverInfo = ServerInfo{}
	result, err := RunAdminCommand(session, "serverStatus")
	if err != nil {
		return serverInfo, err
	}
	bytes, _ := json.Marshal(result)
	stat := ServerStatusDoc{}
	json.Unmarshal(bytes, &stat)

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

	names, _ := session.DatabaseNames()
	dbStats := DBStats{}
	var dataSize, indexSize int
	list := []bson.M{}

	for _, name := range names {
		result, _ = RunCommandOnDB(session, "dbStats", name)
		b, _ := json.Marshal(result)
		json.Unmarshal(b, &dbStats)
		dataSize += dbStats.DataSize
		indexSize += dbStats.IndexSize
		list = append(list, bson.M{"db": name, "objects": dbStats.Objects, "dataSize": dbStats.DataSize, "indexSize": dbStats.IndexSize})
	}

	serverInfo.StorageSize = bson.M{"totalDataSize (MB)": dataSize / 1024 / 1024, "totalIndexSize (MB)": indexSize / 1024 / 1024, "statsDetails": list}
	return serverInfo, nil
}
