// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"encoding/json"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// MongoServerInfo constains server info from db.serverStatus()
type MongoServerInfo struct {
	Cluster     string      `json:"cluster" bson:"cluster"`
	Host        string      `json:"host" bson:"host"`
	Process     string      `json:"process" bson:"process"`
	Version     string      `json:"version" bson:"version"`
	Sharding    interface{} `json:"sharding" bson:"sharding"`
	Repl        interface{} `json:"repl" bson:"repl"`
	StorageSize bson.M
}

// DBStats contains info from db.stats()
type DBStats struct {
	DB        string `json:"db" bson:"db"`
	DataSize  int    `json:"dataSize" bson:"dataSize"`
	IndexSize int    `json:"indexSize" bson:"indexSize"`
	Objects   int    `json:"objects" bson:"objects"`
}

// IsMaster executes dbisMaster()
func IsMaster(session *mgo.Session) (bson.M, error) {
	return AdminCommand(session, "isMaster")
}

// GetMongoServerInfo returns MongoServerInfo from db.serverStatus()
func GetMongoServerInfo(session *mgo.Session) (MongoServerInfo, error) {
	var err error
	info := MongoServerInfo{}
	result, err := AdminCommand(session, "serverStatus")
	if err != nil {
		return info, err
	}
	bytes, _ := json.Marshal(result)
	stat := ServerStatusDoc{}
	json.Unmarshal(bytes, &stat)

	info.Host = stat.Host
	info.Process = stat.Process
	info.Version = stat.Version
	info.Sharding = bson.M{}
	if stat.Sharding != nil {
		info.Sharding = stat.Sharding
	}
	info.Repl = bson.M{}
	if stat.Repl != nil {
		info.Repl = stat.Repl
	}

	if stat.Process == "mongos" {
		info.Cluster = SHARDED
	} else if stat.Repl != nil {
		info.Cluster = "replica"
	} else {
		info.Cluster = STANDALONE
	}

	names, _ := session.DatabaseNames()
	dbStats := DBStats{}
	var dataSize, indexSize int
	list := []bson.M{}

	for _, name := range names {
		result, _ = AdminCommandOnDB(session, "dbStats", name)
		b, _ := json.Marshal(result)
		json.Unmarshal(b, &dbStats)
		dataSize += dbStats.DataSize
		indexSize += dbStats.IndexSize
		list = append(list, bson.M{"db": name, "objects": dbStats.Objects, "dataSize": dbStats.DataSize, "indexSize": dbStats.IndexSize})
	}

	info.StorageSize = bson.M{"totalDataSize (MB)": dataSize / 1024 / 1024, "totalIndexSize (MB)": indexSize / 1024 / 1024, "statsDetails": list}
	return info, nil
}
