package stats

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// ServerStatusInfo -
type ServerStatusInfo struct {
	Cluster      string      `json:"cluster",bson:"cluster"`
	Host         string      `json:"host",bson:"host"`
	Process      string      `json:"process",bson:"process"`
	Version      string      `json:"version",bson:"version"`
	Sharding     interface{} `json:"sharding",bson:"sharding"`
	Repl         interface{} `json:"repl",bson:"repl"`
	TotalDBStats bson.M
}

// DBStats -
type DBStats struct {
	DB        string `json:"db",bson:"db"`
	DataSize  int    `json:"dataSize",bson:"dataSize"`
	IndexSize int    `json:"indexSize",bson:"indexSize"`
}

// GetSession -
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

// IsMaster - Execute isMaster
func IsMaster(session *mgo.Session) bson.M {
	return AdminCommand(session, "isMaster")
}

// ServerInfo -
func ServerInfo(session *mgo.Session) ServerStatusInfo {
	result := AdminCommand(session, "serverStatus")
	bytes, _ := json.Marshal(result)
	stat := ServerStatusData{}
	json.Unmarshal(bytes, &stat)
	ssi := ServerStatusInfo{}

	ssi.Host = stat.Host
	ssi.Process = stat.Process
	ssi.Version = stat.Version
	ssi.Sharding = bson.M{}
	if stat.Sharding != nil {
		ssi.Sharding = stat.Sharding
	}
	ssi.Repl = bson.M{}
	if stat.Repl != nil {
		ssi.Repl = stat.Repl
	}

	if stat.Process == "mongos" {
		ssi.Cluster = "sharded"
	} else if stat.Repl != nil {
		ssi.Cluster = "replica"
	} else {
		ssi.Cluster = "standalone"
	}

	names, _ := session.DatabaseNames()
	dbStats := DBStats{}
	var dsize, isize int
	list := []bson.M{}

	for _, name := range names {
		result = AdminCommandOnDB(session, "dbStats", name)
		bytes, _ := json.Marshal(result)
		json.Unmarshal(bytes, &dbStats)
		dsize += dbStats.DataSize
		isize += dbStats.IndexSize
		list = append(list, bson.M{"db": dbStats.DB, "dataSize": dbStats.DataSize, "indexSize": dbStats.IndexSize})
	}

	ssi.TotalDBStats = bson.M{"totalDataSize": dsize, "totalIndexSize": isize, "statsDetails": list}
	return ssi
}

// AdminCommand - Execute Admin Command
func AdminCommand(session *mgo.Session, command string) bson.M {
	return AdminCommandOnDB(session, command, "admin")
}

// AdminCommandOnDB - Execute Admin Command
func AdminCommandOnDB(session *mgo.Session, command string, db string) bson.M {
	session.SetMode(mgo.Monotonic, true)
	result := bson.M{}
	if err := session.DB(db).Run(command, &result); err != nil {
		fmt.Println(err)
	}
	return result
}
