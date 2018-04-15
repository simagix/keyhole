package mongo

import (
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// IsMaster -
func IsMaster(uri string) bson.M {
	return Admin(uri, "admin", "isMaster")
}

// ServerStatus -
func ServerStatus(uri string) bson.M {
	return Admin(uri, "admin", "serverStatus")
}

// DBStats -
func DBStats(uri string, dbname string) bson.M {
	return Admin(uri, dbname, "dbStats")
}

// Admin -
func Admin(uri string, dbname string, cmd string) bson.M {
	session, err := mgo.Dial(uri)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	result := bson.M{}
	if err := session.DB(dbname).Run(bson.D{{cmd, 1}}, &result); err != nil {
		panic(err)
	}
	return result
}
