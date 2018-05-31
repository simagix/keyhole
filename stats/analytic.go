package stats

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const dateFormat = time.RFC3339

var statsFile = os.TempDir() + "/keyhole_stats." + time.Now().Format("2018-01-02T15-04-05")
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
func (m MongoConn) CollectServerStatus(uri string) {
	pstat := ServerStatusData{}
	stat := ServerStatusData{}
	var iop int
	var piop int

	for {
		session, err := GetSession(uri, m.ssl, m.sslCA)
		if err == nil {
			session.SetMode(mgo.Primary, true)
			serverStatus := m.serverStatus(session)
			bytes, _ := json.Marshal(serverStatus)
			json.Unmarshal(bytes, &stat)
			key := time.Now().Format(dateFormat)
			serverStatusDocs = append(serverStatusDocs, serverStatus)
			if len(serverStatusDocs) > 10 {
				saveStatsToFile()
			}
			mongoStats[key] = stat
			fmt.Printf("\n%s Memory - resident: %7d, virtual: %7d", key, stat.Mem.Resident, stat.Mem.Virtual)
			iop = stat.Metrics.Document.Inserted + stat.Metrics.Document.Returned +
				stat.Metrics.Document.Updated + stat.Metrics.Document.Deleted
			iops := float64(iop-piop) / 60
			if len(serverStatusDocs) > 1 {
				bytes, _ = json.Marshal(serverStatusDocs[len(serverStatusDocs)-2])
				json.Unmarshal(bytes, &pstat)
				fmt.Printf(", page faults: %3d, iops: %7.1f\n", (stat.ExtraInfo.PageFaults - pstat.ExtraInfo.PageFaults), iops)
				fmt.Printf("%s CRUD+  - insert:%7d, find:%7d, update:%7d, delete:%7d, getmore:%7d, command:%7d\n",
					key, stat.OpCounters.Insert-pstat.OpCounters.Insert,
					stat.OpCounters.Query-pstat.OpCounters.Query,
					stat.OpCounters.Update-pstat.OpCounters.Update,
					stat.OpCounters.Delete-pstat.OpCounters.Delete,
					stat.OpCounters.Getmore-pstat.OpCounters.Getmore,
					stat.OpCounters.Command-pstat.OpCounters.Command)
				fmt.Printf("%s Latency- read: %7.1f, write: %7.1f, command: %7.1f (ms)\n",
					key,
					float64(stat.OpLatencies.Reads.Latency-pstat.OpLatencies.Reads.Latency)/float64(stat.OpLatencies.Reads.Ops-pstat.OpLatencies.Reads.Ops)/1000,
					float64(stat.OpLatencies.Writes.Latency-pstat.OpLatencies.Writes.Latency)/float64(stat.OpLatencies.Writes.Ops-pstat.OpLatencies.Writes.Ops)/1000,
					float64(stat.OpLatencies.Commands.Latency-pstat.OpLatencies.Commands.Latency)/float64(stat.OpLatencies.Commands.Ops-pstat.OpLatencies.Commands.Ops)/1000)
			} else {
				fmt.Println()
			}
			piop = iop
			session.Close()
		}
		time.Sleep(1 * time.Minute)
	}
}

// PrintDBStats - Print dbStats every 10 seconds
func (m MongoConn) PrintDBStats() {
	var docs map[string]interface{}
	var pds float64
	var ds float64
	ptime := time.Now()
	now := ptime

	for {
		session, err := GetSession(m.uri, m.ssl, m.sslCA)
		if err == nil {
			session.SetMode(mgo.Primary, true)
			stat := m.dbStats(session)
			bytes, _ := json.Marshal(stat)
			json.Unmarshal(bytes, &docs)
			ds = docs["dataSize"].(float64)
			sec := now.Sub(ptime).Seconds()
			delta := (ds - pds) / mb / sec
			if sec > 1 && delta > .01 {
				fmt.Printf("%s Storage: %6.1f -> %6.1f, rate %6.1f MB/sec\n",
					now.Format(dateFormat), pds/mb, ds/mb, delta)
			}
			pds = ds
			ptime = now
			now = time.Now()
			session.Close()
		}
		time.Sleep(sleepTime)
	}
}

// PrintServerStatus - Print serverStatus summary for the duration
func (m MongoConn) PrintServerStatus(uri string) {
	session, err := GetSession(uri, m.ssl, m.sslCA)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Primary, true)

	serverStatus := m.serverStatus(session)
	stat := ServerStatusData{}
	bytes, _ := json.Marshal(serverStatus)
	json.Unmarshal(bytes, &serverStatus)
	serverStatusDocs = append(serverStatusDocs, serverStatus)
	json.Unmarshal(bytes, &stat)
	mongoStats[time.Now().Format(dateFormat)] = stat

	// save serverStatusDocs
	saveStatsToFile()
	AnalyzeServerStatus(statsFile)
}

func saveStatsToFile() {
	bytes, _ := json.Marshal(serverStatusDocs)
	serverStatusDocs = serverStatusDocs[:0]
	fmt.Println("\nstats written to", statsFile)
	f, ferr := os.OpenFile(statsFile, os.O_WRONLY|os.O_APPEND, 0644)
	if ferr != nil {
		f, _ = os.Create(statsFile)
	}
	defer f.Close()
	f.Write(bytes)
	f.WriteString("\n")
	f.Sync()
	serverStatusDocs = serverStatusDocs[:0]
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
	fmt.Println("filename", filename)
	var serverStatusData = []ServerStatusData{}
	var docs = []ServerStatusData{}
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println("error opening file ", err)
		return
	}
	defer f.Close()

	r := bufio.NewReader(f)
	for {
		text, ferr := r.ReadString('\n') // 0x0A separator = newline
		if ferr == io.EOF {
			break
		}
		json.Unmarshal([]byte(text), &docs)
		serverStatusData = append(serverStatusData, docs...)
	}

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
				stat2.ExtraInfo.PageFaults-stat1.ExtraInfo.PageFaults,
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
	fmt.Println("\n--- Latencies Summary (ms) ---")
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
				stat2.LocalTime.In(loc).Format(dateFormat), r/1000, w/1000, c/1000)
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
