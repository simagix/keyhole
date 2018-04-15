package mongo

import (
	"fmt"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// IsMaster - Execute isMaster
func IsMaster(uri string) bson.M {
	return Admin(uri, "admin", "isMaster")
}

// ServerStatus - Execute serverStatus
func ServerStatus(uri string) bson.M {
	return Admin(uri, "admin", "serverStatus")
}

// DBStats - Execute dbStats
func DBStats(uri string, dbname string) bson.M {
	return Admin(uri, dbname, "dbStats")
}

// Admin - Execute an admin command
func Admin(uri string, dbname string, cmd string) bson.M {
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
