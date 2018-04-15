package mongo

import (
	"bytes"
	"log"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var collname = "keyhole"

// Insert - Insert docs to evaluate performance/bandwidth
func Insert(uri string, dbname string, tps int) {
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
		bt := time.Now()
		for i := 0; i < tps; i++ {
			err = c.Insert(bson.M{"buffer": buffer.String(), "ts": time.Now()})
			if err != nil {
				log.Fatal(err)
			}
		}
		t := time.Now()
		elapsed := t.Sub(bt)
		time.Sleep(time.Second - elapsed)
	}
}
