package stats

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const dateFormat = time.RFC3339

var loc, _ = time.LoadLocation("Local")

var mb = 1024.0 * 1024
var serverStatusDocs = []bson.M{}
var mongoStats = make(map[string]ServerStatusData)
var sleepTime = 10 * time.Second

// simulation -
type analytic struct {
	uri    string
	dbName string
}

// Cache -
type Cache struct {
	MaxBytesConfigured     int `json:"maximum bytes configured" bson:"maximum bytes configured"`
	CurrentlyInCache       int `json:"bytes currently in the cache" bson:"bytes currently in the cache"`
	UnmodifiedPagesEvicted int `json:"unmodified pages evicted" bson:"unmodified pages evicted"`
	TrackedDirtyBytes      int `json:"tracked dirty bytes in the cache" bson:"tracked dirty bytes in the cache"`
	PagesReadIntoCache     int `json:"pages read into cache" bson:"pages read into cache"`
	PagesWrittenFromCache  int `json:"pages written from cache" bson:"pages written from cache"`
}

// WiredTigerData -
type WiredTigerData struct {
	Perf  interface{} `json:"perf" bson:"perf"`
	Cache Cache       `json:"cache" bson:"cache"`
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
	Sharding    interface{}    `json:"sharding" bson:"sharding"`
	Repl        interface{}    `json:"repl" bson:"repl"`
	Process     string         `json:"process" bson:"process"`
	Version     string         `json:"version" bson:"version"`
}

// CollectServerStatus - Collect serverStatus every 10 minutes
func (m MongoConn) CollectServerStatus() {
	session, err := GetSession(m.uri, m.ssl, m.sslCA)
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
		serverStatusDocs = append(serverStatusDocs, serverStatus)
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
	session, err := GetSession(m.uri, m.ssl, m.sslCA)
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
	session, err := GetSession(m.uri, m.ssl, m.sslCA)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	serverStatus := m.serverStatus(session)
	stat := ServerStatusData{}
	bytes, _ := json.Marshal(serverStatus)
	json.Unmarshal(bytes, &serverStatus)
	serverStatusDocs = append(serverStatusDocs, serverStatus)
	json.Unmarshal(bytes, &stat)
	mongoStats[time.Now().Format(dateFormat)] = stat

	// Save mongoStats
	bytes, _ = json.Marshal(serverStatusDocs)
	statsFile := os.TempDir() + "/keyhole_stats." + time.Now().Format("2018-01-02T15-04-05")
	fmt.Println("\nServer status written to", statsFile)
	f, ferr := os.Create(statsFile)
	if ferr != nil {
		panic(ferr)
	}
	defer f.Close()
	f.Write(bytes)
	f.Sync()
	AnalyzeServerStatus(statsFile)
}

// serverStatus - Execute serverStatus
func (m MongoConn) serverStatus(session *mgo.Session) bson.M {
	result := bson.M{}
	session.DB("admin").Run("serverStatus", &result)
	return result
}

// dbStats - Execute dbStats
func (m MongoConn) dbStats(session *mgo.Session) bson.M {
	result := bson.M{}
	session.DB(m.dbName).Run("dbStats", &result)
	return result
}

// AnalyzeServerStatus -
func AnalyzeServerStatus(filename string) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	var serverStatusData = []ServerStatusData{}
	json.Unmarshal(bytes, &serverStatusData)
	PrintStatsDetails(serverStatusData)
	PrintLatencyDetails(serverStatusData)
	PrintMetricsDetails(serverStatusData)
	PrintWiredTigerDetails(serverStatusData)
}

// PrintStatsDetails -
func PrintStatsDetails(docs []ServerStatusData) {
	stat1 := ServerStatusData{}
	stat2 := ServerStatusData{}
	cnt := 0
	fmt.Println("\n--- Analytic Summary ---")
	fmt.Printf("+-------------------------+-------+-------+------+--------+--------+--------+--------+--------+--------+--------+\n")
	fmt.Printf("| Date/Time               | res   | virt  | fault| Command| Delete | Getmore| Insert | Query  | Update | iops   |\n")
	fmt.Printf("|-------------------------|-------+-------|------|--------|--------|--------|--------|--------|--------|--------|\n")
	for _, doc := range docs {
		bytes, _ := json.Marshal(doc)
		json.Unmarshal(bytes, &stat2)
		if cnt > 0 {
			d := int(stat2.LocalTime.Sub(stat1.LocalTime).Seconds())
			iops := stat2.OpCounters.Command - stat1.OpCounters.Command +
				stat2.OpCounters.Delete - stat1.OpCounters.Delete +
				stat2.OpCounters.Getmore - stat1.OpCounters.Getmore +
				stat2.OpCounters.Insert - stat1.OpCounters.Insert +
				stat2.OpCounters.Query - stat1.OpCounters.Query +
				stat2.OpCounters.Update - stat1.OpCounters.Update
			if d > 0 {
				iops = iops / d
			} else {
				iops = 0
			}

			fmt.Printf("|%-25s|%7d|%7d|%6d|%8d|%8d|%8d|%8d|%8d|%8d|%8d|\n",
				stat2.LocalTime.In(loc).Format(dateFormat),
				stat2.Mem.Resident,
				stat2.Mem.Virtual,
				stat2.ExtraInfo.PageFaults,
				stat2.OpCounters.Command-stat1.OpCounters.Command,
				stat2.OpCounters.Delete-stat1.OpCounters.Delete,
				stat2.OpCounters.Getmore-stat1.OpCounters.Getmore,
				stat2.OpCounters.Insert-stat1.OpCounters.Insert,
				stat2.OpCounters.Query-stat1.OpCounters.Query,
				stat2.OpCounters.Update-stat1.OpCounters.Update, iops)
		}
		stat1 = stat2
		cnt++
	}
	fmt.Printf("+-------------------------+-------+-------+------+--------+--------+--------+--------+--------+--------+--------+\n")
}

// PrintLatencyDetails -
func PrintLatencyDetails(docs []ServerStatusData) {
	stat1 := ServerStatusData{}
	stat2 := ServerStatusData{}
	cnt := 0
	fmt.Println("\n--- Latencies Summary ---")
	fmt.Printf("+-------------------------+----------+----------+----------+\n")
	fmt.Printf("| Date/Time               | reads    | writes   | commands |\n")
	fmt.Printf("|-------------------------|----------|----------|----------|\n")
	for _, doc := range docs {
		bytes, _ := json.Marshal(doc)
		json.Unmarshal(bytes, &stat2)
		if cnt > 0 {
			r := stat2.OpLatencies.Reads.Ops - stat1.OpLatencies.Reads.Ops
			if r > 0 {
				r = (stat2.OpLatencies.Reads.Latency - stat1.OpLatencies.Reads.Latency) / r
			}
			w := stat2.OpLatencies.Writes.Ops - stat1.OpLatencies.Writes.Ops
			if w > 0 {
				w = (stat2.OpLatencies.Writes.Latency - stat1.OpLatencies.Writes.Latency) / w
			}
			c := stat2.OpLatencies.Commands.Ops - stat1.OpLatencies.Commands.Ops
			if c > 0 {
				c = (stat2.OpLatencies.Commands.Latency - stat1.OpLatencies.Commands.Latency) / c
			}
			fmt.Printf("|%-25s|%10d|%10d|%10d|\n",
				stat2.LocalTime.In(loc).Format(dateFormat), r, w, c)
		}
		stat1 = stat2
		cnt++
	}
	fmt.Printf("+-------------------------+----------+----------+----------+\n")
}

// PrintMetricsDetails -
func PrintMetricsDetails(docs []ServerStatusData) {
	stat1 := ServerStatusData{}
	stat2 := ServerStatusData{}
	cnt := 0
	fmt.Println("\n--- Metrics ---")
	fmt.Printf("+-------------------------+----------+------------+------------+--------------+----------+----------+----------+----------+\n")
	fmt.Printf("| Date/Time               | Scanned  | ScannedObj |ScanAndOrder|WriteConflicts| Deleted  | Inserted | Returned | Updated  |\n")
	fmt.Printf("|-------------------------|----------|------------|------------|--------------|----------|----------|----------|----------|\n")
	for _, doc := range docs {
		bytes, _ := json.Marshal(doc)
		json.Unmarshal(bytes, &stat2)
		if cnt > 0 {
			fmt.Printf("|%-25s|%10d|%12d|%12d|%14d|%10d|%10d|%10d|%10d|\n",
				stat2.LocalTime.In(loc).Format(dateFormat),
				stat2.Metrics.QueryExecutor.Scanned-stat1.Metrics.QueryExecutor.Scanned,
				stat2.Metrics.QueryExecutor.ScannedObjects-stat1.Metrics.QueryExecutor.ScannedObjects,
				stat2.Metrics.Operation.ScanAndOrder-stat1.Metrics.Operation.ScanAndOrder,
				stat2.Metrics.Operation.WriteConflicts-stat1.Metrics.Operation.WriteConflicts,
				stat2.Metrics.Document.Deleted-stat1.Metrics.Document.Deleted,
				stat2.Metrics.Document.Inserted-stat1.Metrics.Document.Inserted,
				stat2.Metrics.Document.Returned-stat1.Metrics.Document.Returned,
				stat2.Metrics.Document.Updated-stat1.Metrics.Document.Updated)
		}
		stat1 = stat2
		cnt++
	}
	fmt.Printf("+-------------------------+----------+------------+------------+--------------+----------+----------+----------+----------+\n")
}

// PrintWiredTigerDetails -
func PrintWiredTigerDetails(docs []ServerStatusData) {
	stat1 := ServerStatusData{}
	stat2 := ServerStatusData{}
	cnt := 0
	fmt.Println("\n--- WiredTiger Summary ---")
	fmt.Printf("+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+\n")
	fmt.Printf("|                         | MaxBytes     | Currently    | Unmodified   | Tracked      | PagesRead    | PagesWritten |\n")
	fmt.Printf("| Date/Time               | Configured   | InCache      | PagesEvicted | DirtyBytes   | IntoCache    | FromCache    |\n")
	fmt.Printf("|-------------------------|--------------|--------------|--------------|--------------|--------------|--------------|\n")
	for _, doc := range docs {
		bytes, _ := json.Marshal(doc)
		json.Unmarshal(bytes, &stat2)
		if cnt > 0 {
			fmt.Printf("|%-25s|%14d|%14d|%14d|%14d|%14d|%14d|\n",
				stat2.LocalTime.In(loc).Format(dateFormat),
				stat2.WiredTiger.Cache.MaxBytesConfigured,
				stat2.WiredTiger.Cache.CurrentlyInCache,
				stat2.WiredTiger.Cache.UnmodifiedPagesEvicted,
				stat2.WiredTiger.Cache.TrackedDirtyBytes,
				stat2.WiredTiger.Cache.PagesReadIntoCache-stat1.WiredTiger.Cache.PagesReadIntoCache,
				stat2.WiredTiger.Cache.PagesWrittenFromCache-stat1.WiredTiger.Cache.PagesWrittenFromCache)
		}
		stat1 = stat2
		cnt++
	}
	fmt.Printf("+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+\n")
}
