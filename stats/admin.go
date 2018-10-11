// Copyright 2018 Kuei-chun Chen. All rights reserved.

package stats

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"strings"

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

// ParseDialInfo supports seedlist connection string mongodb+srv://
func ParseDialInfo(uri string) (*mgo.DialInfo, error) {
	isSRV := false
	if strings.Index(uri, "mongodb+srv://") == 0 {
		isSRV = true
		// *ssl = true
		uri = "mongodb://" + (uri)[14:]
		if strings.Index(uri, "ssl=") < 0 {
			if strings.Index(uri, "?") < 0 {
				uri = uri + "?ssl=true"
			} else {
				uri = uri + "&ssl=true"
			}
		}
	}

	dialInfo, err := mgo.ParseURL(uri)
	if err != nil {
		return dialInfo, err
	}

	if isSRV == true {
		srvAddr := dialInfo.Addrs[0]
		params, pe := net.LookupTXT(srvAddr)
		if pe != nil {
			fmt.Println("Error:", pe)
			fmt.Println("dialInfo.Addrs:", dialInfo.Addrs)
			return nil, pe
		}
		if strings.Index(uri, "?") < 0 {
			uri = uri + "?" + params[0]
		} else {
			uri = uri + "&" + params[0]
		}
		fmt.Println("uri:", uri)
		dialInfo, err = mgo.ParseURL(uri)
		if err != nil {
			fmt.Println("Error:", err)
			fmt.Println("uri:", uri)
			return dialInfo, err
		}
		_, addrs, le := net.LookupSRV("mongodb", "tcp", srvAddr)
		if le != nil {
			fmt.Println("Error:", le)
			fmt.Println("dialInfo.Addrs:", dialInfo.Addrs)
			return nil, le
		}
		addresses := make([]string, len(addrs))
		for i, addr := range addrs {
			address := strings.TrimSuffix(addr.Target, ".")
			addresses[i] = fmt.Sprintf("%s:%d", address, addr.Port)
		}
		dialInfo.Addrs = addresses
	}
	return dialInfo, err
}

// GetSession returns a MongoDB session
func GetSession(dialInfo *mgo.DialInfo, ssl bool, sslCA string, sslPEMKeyFile string) (*mgo.Session, error) {
	if ssl {
		tlsConfig := &tls.Config{}
		tlsConfig.InsecureSkipVerify = true
		if sslPEMKeyFile != "" {
			clientCertPEM, _ := ioutil.ReadFile(sslPEMKeyFile)
			clientKeyPEM, _ := ioutil.ReadFile(sslPEMKeyFile)
			clientCert, _ := tls.X509KeyPair(clientCertPEM, clientKeyPEM)
			clientCert.Leaf, _ = x509.ParseCertificate(clientCert.Certificate[0])
			tlsConfig.Certificates = []tls.Certificate{clientCert}
		}

		if sslCA != "" {
			var roots *x509.CertPool
			if ca, ferr := ioutil.ReadFile(sslCA); ferr == nil {
				roots = x509.NewCertPool()
				roots.AppendCertsFromPEM(ca)
			}
			tlsConfig.RootCAs = roots
		}

		dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			conn, derr := tls.Dial("tcp", addr.String(), tlsConfig)
			if derr != nil {
				panic(derr)
			}
			return conn, derr
		}
	}
	// dialInfo.Timeout = time.Duration(10 * time.Second)
	return mgo.DialWithInfo(dialInfo)
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
		list = append(list, bson.M{"db": dbStats.DB, "objects": dbStats.Objects, "dataSize": dbStats.DataSize, "indexSize": dbStats.IndexSize})
	}

	info.StorageSize = bson.M{"totalDataSize (MB)": dataSize / 1024 / 1024, "totalIndexSize (MB)": indexSize / 1024 / 1024, "statsDetails": list}
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
