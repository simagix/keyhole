// Copyright 2018 Kuei-chun Chen. All rights reserved.

package stats

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"

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
}

// GetSession returns a MongoDB session
func GetSession(uri string, ssl bool, sslCA string) (*mgo.Session, error) {
	var session *mgo.Session
	var err error

	if ssl {
		roots := x509.NewCertPool()
		if ca, ferr := ioutil.ReadFile(sslCA); ferr == nil {
			roots.AppendCertsFromPEM(ca)
		}
		tlsConfig := &tls.Config{}
		tlsConfig.RootCAs = roots
		dialInfo, perr := mgo.ParseURL(uri)
		if perr != nil {
			panic(perr)
		}
		dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			tlsConfig := &tls.Config{}
			conn, derr := tls.Dial("tcp", addr.String(), tlsConfig)
			if derr != nil {
				panic(derr)
			}
			return conn, derr
		}
		session, err = mgo.DialWithInfo(dialInfo)
	} else {
		session, err = mgo.Dial(uri)
	}

	return session, err
}

// IsMaster executes dbisMaster()
func IsMaster(session *mgo.Session) bson.M {
	return AdminCommand(session, "isMaster")
}

// GetMongoServerInfo returns MongoServerInfo from db.serverStatus()
func GetMongoServerInfo(session *mgo.Session) MongoServerInfo {
	result := AdminCommand(session, "serverStatus")
	bytes, _ := json.Marshal(result)
	stat := ServerStatusDoc{}
	json.Unmarshal(bytes, &stat)
	info := MongoServerInfo{}

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
		info.Cluster = "sharded"
	} else if stat.Repl != nil {
		info.Cluster = "replica"
	} else {
		info.Cluster = "standalone"
	}

	names, _ := session.DatabaseNames()
	dbStats := DBStats{}
	var dataSize, indexSize int
	list := []bson.M{}

	for _, name := range names {
		result = AdminCommandOnDB(session, "dbStats", name)
		bytes, _ := json.Marshal(result)
		json.Unmarshal(bytes, &dbStats)
		dataSize += dbStats.DataSize
		indexSize += dbStats.IndexSize
		list = append(list, bson.M{"db": dbStats.DB, "dataSize": dbStats.DataSize, "indexSize": dbStats.IndexSize})
	}

	info.StorageSize = bson.M{"totalDataSize": dataSize, "totalIndexSize": indexSize, "statsDetails": list}
	return info
}

// AdminCommand executes admin Command
func AdminCommand(session *mgo.Session, command string) bson.M {
	return AdminCommandOnDB(session, command, "admin")
}

// AdminCommandOnDB execute admin Command at given database
func AdminCommandOnDB(session *mgo.Session, command string, db string) bson.M {
	session.SetMode(mgo.Primary, true)
	result := bson.M{}
	if err := session.DB(db).Run(command, &result); err != nil {
		fmt.Println(err)
	}
	return result
}
