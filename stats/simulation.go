package stats

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

var collname = "keyhole"

// MongoConn -
type MongoConn struct {
	uri    string
	dbName string
	tps    int
}

// New - Constructor
func New(uri string, dbName string, tps int) MongoConn {
	m := MongoConn{uri, dbName, tps}
	return m
}

// PopulateData - Insert docs to evaluate performance/bandwidth
func (m MongoConn) PopulateData() {
	session, err := mgo.Dial(m.uri)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	c := session.DB(m.dbName).C(collname)
	var buffer bytes.Buffer
	for i := 0; i < 4096/len("simagix."); i++ {
		buffer.WriteString("simagix.")
	}
	s := 0
	for s < 60 {
		s++
		bt := time.Now()
		for i := 0; i < m.tps; i++ {
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
func (m MongoConn) Simulate(duration int) {
	session, err := mgo.Dial(m.uri)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	c := session.DB(m.dbName).C(collname)
	var buffer bytes.Buffer
	for i := 0; i < 4096/len("simagix."); i++ {
		buffer.WriteString("simagix.")
	}

	result := bson.M{}
	results := []bson.M{}
	change := bson.M{"$set": bson.M{"year": 1989}}
	isBurst := false
	burstBegin := time.NewTimer(2 * time.Minute)
	go func() {
		<-burstBegin.C
		isBurst = true
	}()
	burstEnd := time.NewTimer(time.Duration(duration-2) * time.Minute)
	go func() {
		<-burstEnd.C
		isBurst = false
	}()

	for {
		msec := 5
		if isBurst {
			msec = 1
		}
		id := bson.NewObjectId()
		_ = c.Insert(bson.M{"_id": id, "buffer": buffer.String(), "ts": time.Now()})
		time.Sleep(time.Duration(rand.Intn(msec)) * time.Millisecond)
		_ = c.Find(bson.M{"_id": id}).One(&result)
		time.Sleep(time.Duration(rand.Intn(msec)) * time.Millisecond)
		_ = c.Update(bson.M{"_id": id}, change)
		time.Sleep(time.Duration(rand.Intn(msec)) * time.Millisecond)
		_ = c.Remove(bson.M{"_id": id})
		time.Sleep(time.Duration(rand.Intn(msec)) * time.Millisecond)
		_ = c.Find(nil).Limit(10).All(&results)
		time.Sleep(time.Duration(rand.Intn(msec)) * time.Millisecond)
	}
}

// Cleanup - Drop the temp database
func (m MongoConn) Cleanup() {
	fmt.Println("cleanup", m.uri)
	session, _ := mgo.Dial(m.uri)
	defer session.Close()
	fmt.Println("dropping database", m.dbName)
	time.Sleep(1 * time.Second)
	session.DB(m.dbName).DropDatabase()
}
