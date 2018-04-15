package mongo

import (
	"bytes"
	"log"
	"math/rand"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/simagix/keyhole/stats"

	"github.com/globalsign/mgo"
)

var collname = "keyhole"

// simulation -
type simulation struct {
	uri string
	tps int
}

// New - Constructor
func New(uri string, tps int) simulation {
	sim := simulation{uri, tps}
	return sim
}

// PopulateData - Insert docs to evaluate performance/bandwidth
func (sim simulation) PopulateData() {
	session, err := mgo.Dial(sim.uri)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	c := session.DB(stats.DBName).C(collname)
	var buffer bytes.Buffer
	for i := 0; i < 4096/len("simagix."); i++ {
		buffer.WriteString("simagix.")
	}
	s := 0
	for s < 60 {
		s++
		bt := time.Now()
		for i := 0; i < sim.tps; i++ {
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

// Simulate - Simulate CRUD for load tests
func (sim simulation) Simulate() {
	session, err := mgo.Dial(sim.uri)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	c := session.DB(stats.DBName).C(collname)
	var buffer bytes.Buffer
	for i := 0; i < 4096/len("simagix."); i++ {
		buffer.WriteString("simagix.")
	}

	result := bson.M{}
	results := []bson.M{}
	change := bson.M{"$set": bson.M{"year": 1989}}

	for {
		id := bson.NewObjectId()
		_ = c.Insert(bson.M{"_id": id, "buffer": buffer.String(), "ts": time.Now()})
		time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
		_ = c.Find(bson.M{"_id": id}).One(&result)
		time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
		_ = c.Update(bson.M{"_id": id}, change)
		time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
		_ = c.Remove(bson.M{"_id": id})
		time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
		_ = c.Find(nil).Limit(10).All(&results)
		time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
	}
}
