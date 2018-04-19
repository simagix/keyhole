package stats

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const dateFormat = time.RFC3339

var mb = 1024.0 * 1024
var serverStatusSlice = []bson.M{}
var mongoStats = make(map[string]ServerStatusData)
var sleepTime = 10 * time.Second

// simulation -
type analytic struct {
	uri    string
	dbName string
}

// WiredTigerData -
type WiredTigerData struct {
	Perf  interface{} `json:"perf" bson:"perf"`
	Cache interface{} `json:"cache" bson:"cache"`
}

// OpCounters -
type OpCounters struct {
	Command int `json:"command" bson:"command"`
	Delete  int `json:"delete" bson:"delete"`
	Getmore int `json:"getmore" bson:"getmore"`
	Insert  int `json:"insert" bson:"insert"`
	Query   int `json:"query" bson:"query"`
	Update  int `json:"update" bson:"update"`
}

// OpsLatency -
type OpsLatency struct {
	Latency int `json:"latency" bson:"latency"`
	Ops     int `json:"ops" bson:"ops"`
}

// OpLatencies -
type OpLatencies struct {
	Commands OpsLatency `json:"commands" bson:"commands"`
	Reads    OpsLatency `json:"reads" bson:"reads"`
	Writes   OpsLatency `json:"writes" bson:"writes"`
}

// Document -
type Document struct {
	Deleted  int `json:"deleted" bson:"deleted"`
	Inserted int `json:"inserted" bson:"inserted"`
	Returned int `json:"returned" bson:"returned"`
	Updated  int `json:"updated" bson:"updated"`
}

// QueryExecutor -
type QueryExecutor struct {
	Scanned        int `json:"scanned" bson:"scanned"`
	ScannedObjects int `json:"scannedObjects" bson:"scannedObjects"`
}

// Operation -
type Operation struct {
	ScanAndOrder   int `json:"scanAndOrder" bson:"scanAndOrder"`
	WriteConflicts int `json:"writeConflicts" bson:"writeConflicts"`
}

// Metrics -
type Metrics struct {
	Document      Document      `json:"document" bson:"document"`
	QueryExecutor QueryExecutor `json:"queryExecutor" bson:"queryExecutor"`
	Operation     Operation     `json:"operation" bson:"operation"`
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
	Mem         Mem            `json:"Mem" bson:"Mem"`
	ExtraInfo   ExtraInfo      `json:"extra_info" bson:"extra_info"`
	Metrics     Metrics        `json:"metrics" bson:"metrics"`
	LocalTime   time.Time      `json:"localTime" bson:"localTime"`
	Host        string         `json:"host" bson:"host"`
	OpCounters  OpCounters     `json:"opcounters" bson:"opcounters"`
	OpLatencies OpLatencies    `json:"opLatencies" bson:"opLatencies"`
	WiredTiger  WiredTigerData `json:"wiredTiger" bson:"wiredTiger"`
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
		serverStatus := m.serverStatus(session)
		bytes, _ := json.Marshal(serverStatus)
		json.Unmarshal(bytes, &stat)
		key := time.Now().Format(dateFormat)
		serverStatusSlice = append(serverStatusSlice, serverStatus)
		mongoStats[key] = stat
		fmt.Printf("%s res: %7d, virt: %7d, faults: %5d",
			key, stat.Mem.Resident, stat.Mem.Virtual, stat.ExtraInfo.PageFaults)
		iop = stat.Metrics.Document.Inserted + stat.Metrics.Document.Returned +
			stat.Metrics.Document.Updated + stat.Metrics.Document.Deleted
		iops := (iop - piop) / 60
		if piop > 0 {
			fmt.Printf(", i: %7d, q: %7d, u: %7d, d: %7d, iops: %7d\n",
				stat.Metrics.Document.Inserted-crud[0], stat.Metrics.Document.Returned-crud[1],
				stat.Metrics.Document.Updated-crud[2], stat.Metrics.Document.Deleted-crud[3], iops)
		} else {
			fmt.Println()
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
			fmt.Printf("%s data: %6.1f -> %6.1f, rate %6.1f MB/sec\n",
				now.Format(dateFormat), pds/mb, ds/mb, delta)
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

	serverStatus := m.serverStatus(session)
	stat := ServerStatusData{}
	bytes, _ := json.Marshal(serverStatus)
	json.Unmarshal(bytes, &serverStatus)
	serverStatusSlice = append(serverStatusSlice, serverStatus)
	json.Unmarshal(bytes, &stat)
	mongoStats[time.Now().Format(dateFormat)] = stat

	// Save mongoStats
	bytes, _ = json.Marshal(serverStatusSlice)
	statsFile := os.TempDir() + "keyhole_stats." + time.Now().Format("2018-01-02T15-04-05")
	fmt.Println("\nServer status written to", statsFile)
	f, ferr := os.Create(statsFile)
	if ferr != nil {
		panic(ferr)
	}
	defer f.Close()
	f.Write(bytes)
	f.Sync()

	// Print server status stats
	var keys []string
	for k := range mongoStats {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	key1 := keys[0]
	stat1 := ServerStatusData{}
	bytes, _ = json.Marshal(mongoStats[key1])
	json.Unmarshal(bytes, &stat1)
	// b, _ := json.MarshalIndent(stat1.WiredTiger, "", "  ")
	// fmt.Println("\n", key1, string(b))
	key2 := keys[len(keys)-1]
	stat2 := ServerStatusData{}
	bytes, _ = json.Marshal(mongoStats[key2])
	json.Unmarshal(bytes, &stat2)
	// b, _ = json.MarshalIndent(stat2.WiredTiger, "", "  ")
	// fmt.Println(key2, string(b))
	PrintStatsSummary(stat1, stat2)
}

// PrintStatsSummary -
func PrintStatsSummary(stat1 ServerStatusData, stat2 ServerStatusData) {
	var s1 [6]int
	var s2 [6]int
	s1[0] = stat1.OpCounters.Command
	s1[1] = stat1.OpCounters.Delete
	s1[2] = stat1.OpCounters.Getmore
	s1[3] = stat1.OpCounters.Insert
	s1[4] = stat1.OpCounters.Query
	s1[5] = stat1.OpCounters.Update
	key1 := stat1.LocalTime.Format(dateFormat)
	t1, _ := time.Parse(dateFormat, key1)

	s2[0] = stat2.OpCounters.Command
	s2[1] = stat2.OpCounters.Delete
	s2[2] = stat2.OpCounters.Getmore
	s2[3] = stat2.OpCounters.Insert
	s2[4] = stat2.OpCounters.Query
	s2[5] = stat2.OpCounters.Update
	key2 := stat2.LocalTime.Format(dateFormat)
	t2, _ := time.Parse(dateFormat, key2)
	d := int(t2.Sub(t1).Seconds())
	iops := (s2[0] + s2[1] + s2[2] + s2[3] + s2[4] + s2[5]) - (s1[0] + s1[1] + s1[2] + s1[3] + s1[4] + s1[5])
	iops = iops / d

	// print stats for this duration
	fmt.Println("\n--- Analytic Summary ---")
	fmt.Printf("+-------------------------+-------+-------+------+--------+--------+--------+--------+--------+--------+--------+\n")
	fmt.Printf("| Date/Time               | res   | virt  | fault| Command| Delete | Getmore| Insert | Query  | Update | iops   |\n")
	fmt.Printf("|-------------------------|-------+-------|------|--------|--------|--------|--------|--------|--------|--------|\n")
	fmt.Printf("|%25s|%7d|%7d|%6d|%8d|%8d|%8d|%8d|%8d|%8d|%8s|\n",
		key1, stat1.Mem.Resident, stat1.Mem.Virtual, stat1.ExtraInfo.PageFaults, s1[0], s1[1], s1[2], s1[3], s1[4], s1[5], " ")
	fmt.Printf("|-------------------------|-------+-------|------|--------|--------|--------|--------|--------|--------|--------|\n")
	fmt.Printf("|%25s|%7d|%7d|%6d|%8d|%8d|%8d|%8d|%8d|%8d|%8s|\n",
		key2, stat2.Mem.Resident, stat2.Mem.Virtual, stat2.ExtraInfo.PageFaults, s2[0], s2[1], s2[2], s2[3], s2[4], s2[5], " ")
	fmt.Printf("|-------------------------|-------+-------|------|--------|--------|--------|--------|--------|--------|--------|\n")
	fmt.Printf("|%25s|%7d|%7d|%6d|%8d|%8d|%8d|%8d|%8d|%8d|%8d|\n",
		"Delta",
		stat2.Mem.Resident-stat1.Mem.Resident,
		stat2.Mem.Virtual-stat1.Mem.Virtual,
		stat2.ExtraInfo.PageFaults-stat1.ExtraInfo.PageFaults,
		s2[0]-s1[0], s2[1]-s1[1], s2[2]-s1[2], s2[3]-s1[3], s2[4]-s1[4], s2[5]-s1[5], iops)
	fmt.Printf("+-------------------------+-------+-------+------+--------+--------+--------+--------+--------+--------+--------+\n")
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
