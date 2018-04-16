package stats

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

var mb = 1024.0 * 1024
var mongoStats = make(map[string]interface{})
var sleepTime = 10 * time.Second

// simulation -
type analytic struct {
	uri    string
	dbName string
}

// WiredTigerData -
type WiredTigerData struct {
	Perf interface{}
}

// Document -
type Document struct {
	Deleted  int `json:"deleted" bson:"deleted"`
	Inserted int `json:"inserted" bson:"inserted"`
	Returned int `json:"returned" bson:"returned"`
	Updated  int `json:"updated" bson:"updated"`
}

// Metrics -
type Metrics struct {
	Document Document `json:"document" bson:"document"`
}

// Mem -
type Mem struct {
	Resident int `json:"resident" bson:"resident"`
	Virtual  int `json:"virtual" bson:"virtual"`
}

// ExtraInfo -
type ExtraInfo struct {
	PageFaults int `json:"page_faults" bson:"page_faults"`
}

// ServerStatusData -
type ServerStatusData struct {
	Mem        Mem       `json:"Mem" bson:"Mem"`
	ExtraInfo  ExtraInfo `json:"extra_info" bson:"extra_info"`
	Metrics    Metrics   `json:"metrics" bson:"metrics"`
	WiredTiger WiredTigerData
}

// CollectServerStatus - Collect serverStatus every 10 minutes
func (m MongoConn) CollectServerStatus() {
	session, err := mgo.Dial(m.uri)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	stat := ServerStatusData{}
	var iop int
	var piop int
	var crud [4]int

	for {
		bytes, _ := json.Marshal(m.serverStatus(session))
		json.Unmarshal(bytes, &stat)
		key := time.Now().Format("2018-01-02T15:04:05-05:00")
		mongoStats[key] = stat
		fmt.Printf("%s resident: %d, virtual: %d, page faults: %d\n",
			key, stat.Mem.Resident, stat.Mem.Virtual, stat.ExtraInfo.PageFaults)
		iop = stat.Metrics.Document.Inserted + stat.Metrics.Document.Returned +
			stat.Metrics.Document.Updated + stat.Metrics.Document.Deleted
		iops := (iop - piop) / 60
		if piop > 0 {
			fmt.Printf("%s metrics: c: %d, r: %d, u: %d, d: %d, iops: %d\n",
				key, stat.Metrics.Document.Inserted-crud[0], stat.Metrics.Document.Returned-crud[1],
				stat.Metrics.Document.Updated-crud[2], stat.Metrics.Document.Deleted-crud[3], iops)
		}
		crud[0] = stat.Metrics.Document.Inserted
		crud[1] = stat.Metrics.Document.Returned
		crud[2] = stat.Metrics.Document.Updated
		crud[3] = stat.Metrics.Document.Deleted
		piop = iop
		time.Sleep(1 * time.Minute)
	}
}

// PrintDBStats - Print dbStats every 10 seconds
func (m MongoConn) PrintDBStats() {
	session, err := mgo.Dial(m.uri)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	var raw map[string]interface{}
	var pds float64
	var ds float64
	ptime := time.Now()
	now := ptime

	for {
		stat := m.dbStats(session)
		bytes, _ := json.Marshal(stat)
		json.Unmarshal(bytes, &raw)
		ds = raw["dataSize"].(float64)
		sec := now.Sub(ptime).Seconds()
		delta := (ds - pds) / mb / sec
		if sec > 0 && delta > .01 {
			fmt.Printf("%s %8.1f -> %8.1f, rate %8.1f MB/second\n",
				now.Format("2018-01-02T15:04:05-05:00"), pds/mb, ds/mb, delta)
		}
		pds = ds
		ptime = now
		time.Sleep(sleepTime)
		now = time.Now()
	}
}

// PrintServerStatus - Print serverStatus summary for the duration
func (m MongoConn) PrintServerStatus() {
	session, err := mgo.Dial(m.uri)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	stat := ServerStatusData{}
	bytes, _ := json.Marshal(m.serverStatus(session))
	json.Unmarshal(bytes, &stat)
	mongoStats[time.Now().Format("2018-01-02T15:04:05-05:00")] = stat

	var keys []string
	for k := range mongoStats {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var s1 [4]int
	var s2 [4]int
	key := keys[0]
	stat1 := ServerStatusData{}
	stat2 := ServerStatusData{}
	bytes, _ = json.Marshal(mongoStats[key])
	json.Unmarshal(bytes, &stat1)
	b, _ := json.MarshalIndent(stat1.WiredTiger, "", "  ")
	fmt.Println("\n", key, string(b))
	s1[0] = stat1.Metrics.Document.Inserted
	s1[1] = stat1.Metrics.Document.Returned
	s1[2] = stat1.Metrics.Document.Updated
	s1[3] = stat1.Metrics.Document.Deleted
	t1, _ := time.Parse("2018-01-02T15:04:05-05:00", key)
	key = keys[len(keys)-1]
	bytes, _ = json.Marshal(mongoStats[key])
	json.Unmarshal(bytes, &stat2)
	b, _ = json.MarshalIndent(stat2.WiredTiger, "", "  ")
	fmt.Println(key, string(b))
	s2[0] = stat2.Metrics.Document.Inserted
	s2[1] = stat2.Metrics.Document.Returned
	s2[2] = stat2.Metrics.Document.Updated
	s2[3] = stat2.Metrics.Document.Deleted
	t2, _ := time.Parse("2018-01-02T15:04:05-05:00", key)
	d := int(t2.Sub(t1).Seconds())
	iops := (s2[0] + s2[1] + s2[2] + s2[3]) - (s1[0] + s1[1] + s1[2] + s1[3])
	iops = iops / d

	// print stats for this duration
	fmt.Println("\n--- Analytic Summary ---")
	fmt.Printf("resident mem %d -> %d\n", stat1.Mem.Resident, stat2.Mem.Resident)
	fmt.Printf("virtual mem %d -> %d\n", stat1.Mem.Virtual, stat2.Mem.Virtual)
	fmt.Printf("page faules %d -> %d\n", stat1.ExtraInfo.PageFaults, stat2.ExtraInfo.PageFaults)
	fmt.Printf("%s metrics: c: %d, r: %d, u: %d, d: %d, iops: %d\n",
		key, s2[0]-s1[0], s2[1]-s1[1], s2[2]-s1[2], s2[3]-s1[3], iops)
}

// serverStatus - Execute serverStatus
func (m MongoConn) serverStatus(session *mgo.Session) bson.M {
	result := bson.M{}
	session.DB("admin").Run(bson.D{{"serverStatus", 1}}, &result)
	return result
}

// dbStats - Execute dbStats
func (m MongoConn) dbStats(session *mgo.Session) bson.M {
	result := bson.M{}
	session.DB(m.dbName).Run(bson.D{{"dbStats", 1}}, &result)
	return result
}
