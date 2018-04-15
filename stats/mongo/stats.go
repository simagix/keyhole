package mongo

import (
	"encoding/json"
	"fmt"
	"time"

	mgo "gopkg.in/mgo.v2"
)

var mb = 1024.0 * 1024
var mongoStats = make(map[time.Time]interface{})
var dbStats = make(map[time.Time]interface{})
var sleepTime = 10 * time.Second

// CollectServerStatus -
func CollectServerStatus(uri string) {
	mongoStats[time.Now()] = ServerStatus(uri)

	for {
		time.Sleep(sleepTime)
		mongoStats[time.Now()] = ServerStatus(uri)
	}
}

// CollectDBStats -
func CollectDBStats(uri string, dbname string) {
	var raw map[string]interface{}
	pds := 0.0
	ds := 0.0
	ptime := time.Now()
	now := ptime

	for {
		stat := DBStats(uri, dbname)
		dbStats[now] = stat
		bytes, _ := json.MarshalIndent(stat, "", "   ")
		json.Unmarshal(bytes, &raw)
		ds = raw["dataSize"].(float64)
		sec := now.Sub(ptime).Seconds()
		if sec > 0 {
			fmt.Printf("%s %8.1f -> %8.1f, rate %8.1f MB/second\n",
				now.Format("2006-01-02T15:04:05-07:00"), pds/mb, ds/mb, (ds-pds)/mb/sec)
		}
		pds = ds
		ptime = now
		time.Sleep(sleepTime)
		now = time.Now()
	}
}

// PrintServerStatus -
func PrintServerStatus() {
	for key, value := range mongoStats {
		fmt.Println("Key:", key, "Value:", value)
	}
}

// Cleanup -
func Cleanup(uri string, dbname string) {
	fmt.Println("cleanup", uri)
	session, _ := mgo.Dial(uri)
	defer session.Close()
	fmt.Println("dropping database", dbname)
	time.Sleep(1 * time.Second)
	session.DB(dbname).DropDatabase()
}
