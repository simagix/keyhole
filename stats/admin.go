package stats

import (
	"fmt"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// IsMaster - Execute isMaster
func IsMaster(uri string) bson.M {
	session, err := mgo.Dial(uri)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	result := bson.M{}
	if err := session.DB("admin").Run(bson.D{{"isMaster", 1}}, &result); err != nil {
		fmt.Println(err)
	}
	return result
}
