// Copyright 2018 Kuei-chun Chen. All rights reserved.

package stats

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var keyholeStatsDataFile = os.TempDir() + "/keyhole_stats." + strings.Replace(time.Now().Format(time.RFC3339)[:19], ":", "", -1)
var loc, _ = time.LoadLocation("Local")
var mb = 1024.0 * 1024
var serverStatusDocs = map[string][]bson.M{}

// DocumentDoc contains db.serverStatus().document
type DocumentDoc struct {
	Deleted  int `json:"deleted" bson:"deleted"`
	Inserted int `json:"inserted" bson:"inserted"`
	Returned int `json:"returned" bson:"returned"`
	Updated  int `json:"updated" bson:"updated"`
}

// ExtraInfoDoc contains db.serverStatus().extra_info
type ExtraInfoDoc struct {
	PageFaults int `json:"page_faults" bson:"page_faults"`
}

// MemDoc containers db.serverStatus().mem
type MemDoc struct {
	Resident int `json:"resident" bson:"resident"`
	Virtual  int `json:"virtual" bson:"virtual"`
}

// MetricsDoc contains db.serverStatus().metrics
type MetricsDoc struct {
	Document      DocumentDoc      `json:"document" bson:"document"`
	QueryExecutor QueryExecutorDoc `json:"queryExecutor" bson:"queryExecutor"`
	Operation     OperationDoc     `json:"operation" bson:"operation"`
}

// OperationDoc contains db.serverStatus().operation
type OperationDoc struct {
	ScanAndOrder   int `json:"scanAndOrder" bson:"scanAndOrder"`
	WriteConflicts int `json:"writeConflicts" bson:"writeConflicts"`
}

// OpCountersDoc contains db.serverStatus().OpCounters
type OpCountersDoc struct {
	Command int `json:"command" bson:"command"`
	Delete  int `json:"delete" bson:"delete"`
	Getmore int `json:"getmore" bson:"getmore"`
	Insert  int `json:"insert" bson:"insert"`
	Query   int `json:"query" bson:"query"`
	Update  int `json:"update" bson:"update"`
}

// OpLatenciesDoc contains db.serverStatus().opLatencies
type OpLatenciesDoc struct {
	Commands OpLatenciesOpDoc `json:"commands" bson:"commands"`
	Reads    OpLatenciesOpDoc `json:"reads" bson:"reads"`
	Writes   OpLatenciesOpDoc `json:"writes" bson:"writes"`
}

// OpLatenciesOpDoc contains doc of db.serverStatus().opLatencies
type OpLatenciesOpDoc struct {
	Latency int `json:"latency" bson:"latency"`
	Ops     int `json:"ops" bson:"ops"`
}

// QueryExecutorDoc contains db.serverStatus().queryExecutor
type QueryExecutorDoc struct {
	Scanned        int `json:"scanned" bson:"scanned"`
	ScannedObjects int `json:"scannedObjects" bson:"scannedObjects"`
}

// WiredTigerCacheDoc contains db.serverStatus().wiredTiger.cache
type WiredTigerCacheDoc struct {
	MaxBytesConfigured     int `json:"maximum bytes configured" bson:"maximum bytes configured"`
	CurrentlyInCache       int `json:"bytes currently in the cache" bson:"bytes currently in the cache"`
	UnmodifiedPagesEvicted int `json:"unmodified pages evicted" bson:"unmodified pages evicted"`
	TrackedDirtyBytes      int `json:"tracked dirty bytes in the cache" bson:"tracked dirty bytes in the cache"`
	PagesReadIntoCache     int `json:"pages read into cache" bson:"pages read into cache"`
	PagesWrittenFromCache  int `json:"pages written from cache" bson:"pages written from cache"`
}

// WiredTigerDoc containers db.serverStatus().wiredTiger
type WiredTigerDoc struct {
	Perf  interface{}        `json:"perf" bson:"perf"`
	Cache WiredTigerCacheDoc `json:"cache" bson:"cache"`
}

// ServerStatusDoc contains docs from db.serverStatus()
type ServerStatusDoc struct {
	ExtraInfo   ExtraInfoDoc   `json:"extra_info" bson:"extra_info"`
	Host        string         `json:"host" bson:"host"`
	LocalTime   time.Time      `json:"localTime" bson:"localTime"`
	Mem         MemDoc         `json:"Mem" bson:"Mem"`
	Metrics     MetricsDoc     `json:"metrics" bson:"metrics"`
	OpCounters  OpCountersDoc  `json:"opcounters" bson:"opcounters"`
	OpLatencies OpLatenciesDoc `json:"opLatencies" bson:"opLatencies"`
	Process     string         `json:"process" bson:"process"`
	Repl        interface{}    `json:"repl" bson:"repl"`
	Sharding    interface{}    `json:"sharding" bson:"sharding"`
	Version     string         `json:"version" bson:"version"`
	WiredTiger  WiredTigerDoc  `json:"wiredTiger" bson:"wiredTiger"`
}

// CollectServerStatus collects db.serverStatus() every minute
func (m MongoConn) CollectServerStatus(uri string, channel chan string) {
	fmt.Println("CollectServerStatus: connect to", uri)
	pstat := ServerStatusDoc{}
	stat := ServerStatusDoc{}
	var iop int
	var piop int
	dialInfo, _ := mgo.ParseURL(uri)
	if dialInfo.ReplicaSetName == "" {
		dialInfo.ReplicaSetName = "standalone"
	}

	for {
		session, err := GetSession(uri, m.ssl, m.sslCA)
		if err == nil {
			session.SetMode(mgo.Primary, true)
			serverStatus := m.serverStatus(session)
			bytes, _ := json.Marshal(serverStatus)
			json.Unmarshal(bytes, &stat)
			key := time.Now().Format(time.RFC3339)
			serverStatusDocs[uri] = append(serverStatusDocs[uri], serverStatus)
			if len(serverStatusDocs) > 10 {
				saveServerStatusDocsToFile(uri)
			}
			if dialInfo.ReplicaSetName == "" {
				dialInfo.ReplicaSetName = "standalone"
			}
			str := fmt.Sprintf("\n%s [%s] Memory - resident: %d, virtual: %d",
				key, dialInfo.ReplicaSetName, stat.Mem.Resident, stat.Mem.Virtual)
			iop = stat.Metrics.Document.Inserted + stat.Metrics.Document.Returned +
				stat.Metrics.Document.Updated + stat.Metrics.Document.Deleted
			iops := float64(iop-piop) / 60
			if len(serverStatusDocs[uri]) > 1 {
				bytes, _ = json.Marshal(serverStatusDocs[uri][len(serverStatusDocs[uri])-2])
				json.Unmarshal(bytes, &pstat)
				if stat.Host == pstat.Host {
					str += fmt.Sprintf(", page faults: %d, iops: %.1f\n",
						(stat.ExtraInfo.PageFaults - pstat.ExtraInfo.PageFaults), iops)
					str += fmt.Sprintf("%s [%s] CRUD+  - insert: %d, find: %d, update: %d, delete: %d, getmore: %d, command: %d\n",
						key, dialInfo.ReplicaSetName, stat.OpCounters.Insert-pstat.OpCounters.Insert,
						stat.OpCounters.Query-pstat.OpCounters.Query,
						stat.OpCounters.Update-pstat.OpCounters.Update,
						stat.OpCounters.Delete-pstat.OpCounters.Delete,
						stat.OpCounters.Getmore-pstat.OpCounters.Getmore,
						stat.OpCounters.Command-pstat.OpCounters.Command)
					str += fmt.Sprintf("%s [%s] Latency- read: %.1f, write: %.1f, command: %.1f (ms)\n",
						key, dialInfo.ReplicaSetName,
						float64(stat.OpLatencies.Reads.Latency-pstat.OpLatencies.Reads.Latency)/float64(stat.OpLatencies.Reads.Ops-pstat.OpLatencies.Reads.Ops)/1000,
						float64(stat.OpLatencies.Writes.Latency-pstat.OpLatencies.Writes.Latency)/float64(stat.OpLatencies.Writes.Ops-pstat.OpLatencies.Writes.Ops)/1000,
						float64(stat.OpLatencies.Commands.Latency-pstat.OpLatencies.Commands.Latency)/float64(stat.OpLatencies.Commands.Ops-pstat.OpLatencies.Commands.Ops)/1000)
				} else {
					str += "\n"
				}
			} else {
				str += "\n"
			}
			channel <- str
			piop = iop
			session.Close()
		}
		time.Sleep(1 * time.Minute)
	}
}

// CollectDBStats collects dbStats every 10 seconds
func (m MongoConn) CollectDBStats(uri string, channel chan string) {
	fmt.Println("CollectDBStats: connect to", uri)
	var docs map[string]interface{}
	var prevDataSize float64
	var dataSize float64
	prevTime := time.Now()
	now := prevTime
	dialInfo, _ := mgo.ParseURL(uri)
	if dialInfo.ReplicaSetName == "" {
		dialInfo.ReplicaSetName = "standalone"
	}
	for {
		session, err := GetSession(uri, m.ssl, m.sslCA)
		if err == nil {
			session.SetMode(mgo.Primary, true)
			stat := m.dbStats(session)
			bytes, _ := json.Marshal(stat)
			json.Unmarshal(bytes, &docs)
			if docs["dataSize"] != nil {
				dataSize = docs["dataSize"].(float64)
			}
			sec := now.Sub(prevTime).Seconds()
			delta := (dataSize - prevDataSize) / mb / sec
			if sec > 1 && delta > .01 {
				str := fmt.Sprintf("%s [%s] Storage: %.1f -,> %.1f, rate: %.1f MB/sec\n",
					now.Format(time.RFC3339), dialInfo.ReplicaSetName, prevDataSize/mb, dataSize/mb, delta)
				channel <- str
			}
			prevDataSize = dataSize
			prevTime = now
			now = time.Now()
			session.Close()
		}
		time.Sleep(10 * time.Second)
	}
}

// PrintServerStatus prints serverStatusDocs summary for the duration
func (m MongoConn) PrintServerStatus(uri string) {
	session, err := GetSession(uri, m.ssl, m.sslCA)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Primary, true)

	serverStatus := m.serverStatus(session)
	bytes, _ := json.Marshal(serverStatus)
	json.Unmarshal(bytes, &serverStatus)
	serverStatusDocs[uri] = append(serverStatusDocs[uri], serverStatus)
	filename := saveServerStatusDocsToFile(uri)
	AnalyzeServerStatus(filename)
}

// saveServerStatusDocsToFile appends []ServerStatusDoc to a file
func saveServerStatusDocsToFile(uri string) string {
	dialInfo, _ := mgo.ParseURL(uri)
	if dialInfo.ReplicaSetName == "" {
		dialInfo.ReplicaSetName = "standalone"
	}
	bytes, _ := json.Marshal(serverStatusDocs[uri])
	serverStatusDocs[uri] = serverStatusDocs[uri][:0]
	filename := keyholeStatsDataFile + "-" + dialInfo.ReplicaSetName
	fmt.Println("\nstats written to", filename)
	f, ferr := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND, 0644)
	if ferr != nil {
		f, _ = os.Create(filename)
	}
	defer f.Close()
	f.Write(bytes)
	f.WriteString("\n")
	f.Sync()
	serverStatusDocs[uri] = serverStatusDocs[uri][:0]
	return filename
}

// serverStatus executes db.serverStatus()
func (m MongoConn) serverStatus(session *mgo.Session) bson.M {
	result := bson.M{}
	session.DB("admin").Run("serverStatus", &result)
	return result
}

// dbStats executes db.Stats()
func (m MongoConn) dbStats(session *mgo.Session) bson.M {
	result := bson.M{}
	session.DB(m.dbName).Run("dbStats", &result)
	return result
}

// AnalyzeServerStatus -
func AnalyzeServerStatus(filename string) {
	fmt.Println("filename", filename)
	var allDocs = []ServerStatusDoc{}
	var docs = []ServerStatusDoc{}
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println("error opening file ", err)
		return
	}
	defer f.Close()

	r := bufio.NewReader(f)
	for {
		line, ferr := r.ReadString('\n') // 0x0A separator = newline
		if ferr == io.EOF {
			break
		}
		json.Unmarshal([]byte(line), &docs)
		allDocs = append(allDocs, docs...)
	}

	PrintStatsDetails(allDocs)
	PrintLatencyDetails(allDocs)
	PrintMetricsDetails(allDocs)
	PrintWiredTigerDetails(allDocs)
}

// PrintStatsDetails -
func PrintStatsDetails(docs []ServerStatusDoc) {
	stat1 := ServerStatusDoc{}
	stat2 := ServerStatusDoc{}
	cnt := 0
	fmt.Println("\n--- Analytic Summary ---")
	fmt.Printf("+-------------------------+-------+-------+------+--------+--------+--------+--------+--------+--------+--------+\n")
	fmt.Printf("| Date/Time               | res   | virt  | fault| Command| Delete | Getmore| Insert | Query  | Update | iops   |\n")
	fmt.Printf("|-------------------------|-------+-------|------|--------|--------|--------|--------|--------|--------|--------|\n")
	for _, doc := range docs {
		bytes, _ := json.Marshal(doc)
		json.Unmarshal(bytes, &stat2)
		if cnt > 0 && stat2.Host == stat1.Host {
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
				stat2.LocalTime.In(loc).Format(time.RFC3339),
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
func PrintLatencyDetails(docs []ServerStatusDoc) {
	stat1 := ServerStatusDoc{}
	stat2 := ServerStatusDoc{}
	cnt := 0
	fmt.Println("\n--- Latencies Summary (ms) ---")
	fmt.Printf("+-------------------------+----------+----------+----------+\n")
	fmt.Printf("| Date/Time               | reads    | writes   | commands |\n")
	fmt.Printf("|-------------------------|----------|----------|----------|\n")
	for _, doc := range docs {
		bytes, _ := json.Marshal(doc)
		json.Unmarshal(bytes, &stat2)
		if cnt > 0 && stat2.Host == stat1.Host {
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
				stat2.LocalTime.In(loc).Format(time.RFC3339), r/1000, w/1000, c/1000)
		}
		stat1 = stat2
		cnt++
	}
	fmt.Printf("+-------------------------+----------+----------+----------+\n")
}

// PrintMetricsDetails -
func PrintMetricsDetails(docs []ServerStatusDoc) {
	stat1 := ServerStatusDoc{}
	stat2 := ServerStatusDoc{}
	cnt := 0
	fmt.Println("\n--- Metrics ---")
	fmt.Printf("+-------------------------+----------+------------+------------+--------------+----------+----------+----------+----------+\n")
	fmt.Printf("| Date/Time               | Scanned  | ScannedObj |ScanAndOrder|WriteConflicts| Deleted  | Inserted | Returned | Updated  |\n")
	fmt.Printf("|-------------------------|----------|------------|------------|--------------|----------|----------|----------|----------|\n")
	for _, doc := range docs {
		bytes, _ := json.Marshal(doc)
		json.Unmarshal(bytes, &stat2)
		if cnt > 0 && stat2.Host == stat1.Host {
			fmt.Printf("|%-25s|%10d|%12d|%12d|%14d|%10d|%10d|%10d|%10d|\n",
				stat2.LocalTime.In(loc).Format(time.RFC3339),
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
func PrintWiredTigerDetails(docs []ServerStatusDoc) {
	stat1 := ServerStatusDoc{}
	stat2 := ServerStatusDoc{}
	cnt := 0
	fmt.Println("\n--- WiredTiger Summary ---")
	fmt.Printf("+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+\n")
	fmt.Printf("|                         | MaxBytes     | Currently    | Unmodified   | Tracked      | PagesRead    | PagesWritten |\n")
	fmt.Printf("| Date/Time               | Configured   | InCache      | PagesEvicted | DirtyBytes   | IntoCache    | FromCache    |\n")
	fmt.Printf("|-------------------------|--------------|--------------|--------------|--------------|--------------|--------------|\n")
	for _, doc := range docs {
		bytes, _ := json.Marshal(doc)
		json.Unmarshal(bytes, &stat2)
		if cnt > 0 && stat2.Host == stat1.Host {
			fmt.Printf("|%-25s|%14d|%14d|%14d|%14d|%14d|%14d|\n",
				stat2.LocalTime.In(loc).Format(time.RFC3339),
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
