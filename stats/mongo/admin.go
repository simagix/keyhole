package mongo

import (
	"fmt"
	"time"

	"github.com/simagix/keyhole/stats"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// IsMaster - Execute isMaster
func IsMaster(uri string) bson.M {
	return admin(uri, "admin", "isMaster")
}

// ServerStatus - Execute serverStatus
func ServerStatus(uri string) bson.M {
	return admin(uri, "admin", "serverStatus")
}

// DBStats - Execute dbStats
func DBStats(uri string) bson.M {
	return admin(uri, stats.DBName, "dbStats")
}

// admin - Execute an admin command
func admin(uri string, dbname string, cmd string) bson.M {
	session, err := mgo.Dial(uri)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	result := bson.M{}
	if err := session.DB(dbname).Run(bson.D{{cmd, 1}}, &result); err != nil {
		fmt.Println(err)
	}
	return result
}

// Cleanup - Drop the temp database
func Cleanup(uri string) {
	fmt.Println("cleanup", uri)
	session, _ := mgo.Dial(uri)
	defer session.Close()
	fmt.Println("dropping database", stats.DBName)
	time.Sleep(1 * time.Second)
	session.DB(stats.DBName).DropDatabase()
}
