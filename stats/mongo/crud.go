package mongo

import (
	"bytes"
	"log"
	"math/rand"
	"time"

	"github.com/simagix/keyhole/stats"
	mgo "gopkg.in/mgo.v2"
)

var collname = "keyhole"

// Insert -
func Insert(uri string, dbname string) {
	session, err := mgo.Dial(uri)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	c := session.DB(dbname).C(collname)
	var buffer bytes.Buffer
	for i := 0; i < 4096/len("simagix."); i++ {
		buffer.WriteString("simagix.")
	}
	for {
		year := time.Now().Year() - rand.Intn(5)
		err = c.Insert(&stats.Task{buffer.String(), year})
		if err != nil {
			log.Fatal(err)
		}
	}
}
