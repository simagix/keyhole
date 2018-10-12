// Copyright 2018 Kuei-chun Chen. All rights reserved.

package stats

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/simagix/keyhole/utils"
)

var keyholeStatsDataFile = os.TempDir() + "/keyhole_stats." + strings.Replace(time.Now().Format(time.RFC3339)[:19], ":", "", -1)
var loc, _ = time.LoadLocation("Local")
var mb = 1024.0 * 1024
var serverStatusDocs = map[string][]bson.M{}

// ChartsDocs for drawing charts
var ChartsDocs = map[string][]bson.M{}

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

// GlobalLockSubDoc contains db.serverStatus().globalLockDoc.[activeClients|currentQueue]
type GlobalLockSubDoc struct {
	Readers int `json:"readers" bson:"readers"`
	Total   int `json:"total" bson:"total"`
	Writers int `json:"writers" bson:"writers"`
}

// GlobalLockDoc contains db.serverStatus().globalLockDoc
type GlobalLockDoc struct {
	ActiveClients GlobalLockSubDoc `json:"activeClients" bson:"activeClients"`
	CurrentQueue  GlobalLockSubDoc `json:"currentQueue" bson:"currentQueue"`
	TotalTime     int              `json:"totalTime" bson:"totalTime"`
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
	ModifiedPagesEvicted   int `json:"modified pages evicted" bson:"modified pages evicted"`
	UnmodifiedPagesEvicted int `json:"unmodified pages evicted" bson:"unmodified pages evicted"`
	TrackedDirtyBytes      int `json:"tracked dirty bytes in the cache" bson:"tracked dirty bytes in the cache"`
	PagesReadIntoCache     int `json:"pages read into cache" bson:"pages read into cache"`
	PagesWrittenFromCache  int `json:"pages written from cache" bson:"pages written from cache"`
}

// ConcurrentTransactionsCountDoc contains db.serverStatus().wiredTiger.concurrentTransactions.[read|write]
type ConcurrentTransactionsCountDoc struct {
	Available    int `json:"available" bson:"available"`
	Out          int `json:"out" bson:"out"`
	TotalTickets int `json:"totalTickets" bson:"totalTickets"`
}

// ConcurrentTransactionsDoc contains db.serverStatus().wiredTiger.concurrentTransactions
type ConcurrentTransactionsDoc struct {
	Read  ConcurrentTransactionsCountDoc `json:"read" bson:"read"`
	Write ConcurrentTransactionsCountDoc `json:"write" bson:"write"`
}

// WiredTigerDoc containers db.serverStatus().wiredTiger
type WiredTigerDoc struct {
	Perf                   interface{}               `json:"perf" bson:"perf"`
	Cache                  WiredTigerCacheDoc        `json:"cache" bson:"cache"`
	ConcurrentTransactions ConcurrentTransactionsDoc `json:"concurrentTransactions" bson:"concurrentTransactions"`
}

// ServerStatusDoc contains docs from db.serverStatus()
type ServerStatusDoc struct {
	ExtraInfo   ExtraInfoDoc   `json:"extra_info" bson:"extra_info"`
	GlobalLock  GlobalLockDoc  `json:"globalLock" bson:"globalLock"`
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
	mapKey := getKeyFromReplicaSetName(uri)
	channel <- "CollectServerStatus: connect to " + mapKey + "\n"
	pstat := ServerStatusDoc{}
	stat := ServerStatusDoc{}
	var iop int
	var piop int
	wSeconds := 10
	if m.verbose {
		rstr := fmt.Sprintf("CollectServerStatus collects every %d seconds(s)\n", wSeconds)
		channel <- rstr
	}

	dialInfo, _ := mgo.ParseURL(uri)
	for {
		session, err := GetSession(dialInfo, m.ssl, m.sslCAFile, m.sslPEMKeyFile)
		if err == nil {
			session.SetMode(mgo.Primary, true)
			serverStatus := serverStatus(session)
			buf, _ := json.Marshal(serverStatus)
			json.Unmarshal(buf, &stat)
			key := time.Now().Format(time.RFC3339)
			serverStatusDocs[uri] = append(serverStatusDocs[uri], serverStatus)
			dkey := dialInfo.ReplicaSetName + "/" + strings.Join(dialInfo.Addrs[:], ",")
			ChartsDocs[dkey] = append(ChartsDocs[dkey], serverStatus)
			for len(ChartsDocs[dkey]) > 60 { // shift
				ChartsDocs[dkey] = ChartsDocs[dkey][1:]
			}
			if len(serverStatusDocs[uri]) > 12 {
				saveServerStatusDocsToFile(uri)
			}

			str := fmt.Sprintf("\n%s [%s] Memory - resident: %d, virtual: %d",
				key, mapKey, stat.Mem.Resident, stat.Mem.Virtual)
			iop = stat.Metrics.Document.Inserted + stat.Metrics.Document.Returned +
				stat.Metrics.Document.Updated + stat.Metrics.Document.Deleted
			iops := float64(iop-piop) / 60
			if len(serverStatusDocs[uri]) > 6 && len(serverStatusDocs[uri])%6 == 1 {
				buf, _ = json.Marshal(serverStatusDocs[uri][len(serverStatusDocs[uri])-7])
				json.Unmarshal(buf, &pstat)
				if stat.Host == pstat.Host {
					str += fmt.Sprintf(", page faults: %d, iops: %.1f\n",
						(stat.ExtraInfo.PageFaults - pstat.ExtraInfo.PageFaults), iops)
					str += fmt.Sprintf("%s [%s] CRUD+  - insert: %d, find: %d, update: %d, delete: %d, getmore: %d, command: %d\n",
						key, mapKey, stat.OpCounters.Insert-pstat.OpCounters.Insert,
						stat.OpCounters.Query-pstat.OpCounters.Query,
						stat.OpCounters.Update-pstat.OpCounters.Update,
						stat.OpCounters.Delete-pstat.OpCounters.Delete,
						stat.OpCounters.Getmore-pstat.OpCounters.Getmore,
						stat.OpCounters.Command-pstat.OpCounters.Command)
					str += fmt.Sprintf("%s [%s] Latency- read: %.1f, write: %.1f, command: %.1f (ms)\n",
						key, mapKey,
						float64(stat.OpLatencies.Reads.Latency-pstat.OpLatencies.Reads.Latency)/float64(stat.OpLatencies.Reads.Ops-pstat.OpLatencies.Reads.Ops)/1000,
						float64(stat.OpLatencies.Writes.Latency-pstat.OpLatencies.Writes.Latency)/float64(stat.OpLatencies.Writes.Ops-pstat.OpLatencies.Writes.Ops)/1000,
						float64(stat.OpLatencies.Commands.Latency-pstat.OpLatencies.Commands.Latency)/float64(stat.OpLatencies.Commands.Ops-pstat.OpLatencies.Commands.Ops)/1000)
				} else {
					str += "\n"
				}
			} else {
				str += "\n"
			}
			if m.monitor == false && len(serverStatusDocs[uri])%6 == 1 {
				channel <- str
			}
			piop = iop
			session.Close()
		}
		time.Sleep(time.Duration(wSeconds) * time.Second)
	}
}

// CollectDBStats collects dbStats every 10 seconds
func (m MongoConn) CollectDBStats(uri string, channel chan string, dbName string) {
	mapKey := getKeyFromReplicaSetName(uri)
	channel <- "CollectDBStats: connect to " + mapKey + ", " + dbName + "\n"
	var docs map[string]interface{}
	var prevDataSize float64
	var dataSize float64
	prevTime := time.Now()
	now := prevTime
	dialInfo, _ := mgo.ParseURL(uri)
	session, err := GetSession(dialInfo, m.ssl, m.sslCAFile, m.sslPEMKeyFile)
	defer session.Close()
	for i := 0; i < 10; i++ { // no need to collect after first 1.5 minutes
		if err == nil {
			session.SetMode(mgo.Primary, true)
			stat := dbStats(session, dbName)
			buf, _ := json.Marshal(stat)
			json.Unmarshal(buf, &docs)
			if docs["dataSize"] != nil {
				dataSize = docs["dataSize"].(float64)
			}
			sec := now.Sub(prevTime).Seconds()
			delta := (dataSize - prevDataSize) / mb / sec
			if sec > 1 && delta > .01 {
				str := fmt.Sprintf("%s [%s] Storage: %.1f -> %.1f, rate: %.1f MB/sec\n",
					now.Format(time.RFC3339), mapKey, prevDataSize/mb, dataSize/mb, delta)
				channel <- str
			}
			prevDataSize = dataSize
			prevTime = now
			now = time.Now()
		}
		time.Sleep(10 * time.Second)
	}
	channel <- "CollectDBStats exiting...\n"
}

// PrintServerStatus prints serverStatusDocs summary for the duration
func (m MongoConn) PrintServerStatus(uri string, span int) {
	dialInfo, _ := mgo.ParseURL(uri)
	session, err := GetSession(dialInfo, m.ssl, m.sslCAFile, m.sslPEMKeyFile)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Primary, true)

	serverStatus := serverStatus(session)
	buf, _ := json.Marshal(serverStatus)
	json.Unmarshal(buf, &serverStatus)
	serverStatusDocs[uri] = append(serverStatusDocs[uri], serverStatus)
	filename := saveServerStatusDocsToFile(uri)
	fmt.Println("\nstats written to", filename)
	AnalyzeServerStatus(filename, span, false)
}

// saveServerStatusDocsToFile appends []ServerStatusDoc to a file
func saveServerStatusDocsToFile(uri string) string {
	mapKey := getKeyFromReplicaSetName(uri)
	buf, _ := json.Marshal(serverStatusDocs[uri])
	serverStatusDocs[uri] = serverStatusDocs[uri][:0]
	filename := keyholeStatsDataFile + "-" + mapKey + ".gz"
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	gz.Write(buf)
	gz.Write([]byte{'\n'})
	gz.Close() // close this before flushing the bytes to the buffer.

	f, ferr := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND, 0644)
	if ferr != nil {
		f, _ = os.Create(filename)
	}
	defer f.Close()
	f.Write(b.Bytes())
	f.Sync()
	serverStatusDocs[uri] = serverStatusDocs[uri][:0]
	return filename
}

// serverStatus executes db.serverStatus()
func serverStatus(session *mgo.Session) bson.M {
	result := bson.M{}
	session.DB("admin").Run("serverStatus", &result)
	return result
}

// dbStats executes db.Stats()
func dbStats(session *mgo.Session, dbName string) bson.M {
	result := bson.M{}
	session.DB(dbName).Run("dbStats", &result)
	return result
}

// AnalyzeServerStatus -
func AnalyzeServerStatus(filename string, span int, isWeb bool) {
	// fmt.Println("filename", filename)
	var allDocs = []ServerStatusDoc{}
	var docs = []ServerStatusDoc{}
	var bmap = []bson.M{}
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println("error opening file ", err)
		return
	}
	defer f.Close()

	r := utils.NewReader(f)
	for {
		line, ferr := r.ReadString('\n') // 0x0A separator = newline
		if ferr == io.EOF {
			break
		}
		json.Unmarshal([]byte(line), &docs)
		allDocs = append(allDocs, docs...)
	}

	buf, _ := json.Marshal(allDocs)
	json.Unmarshal(buf, &bmap)
	ChartsDocs["replset"] = bmap
	if isWeb {
		return
	}
	if len(allDocs) > 0 {
		stat := ServerStatusDoc{}
		buf, _ := json.Marshal(allDocs[0])
		json.Unmarshal(buf, &stat)
		fmt.Printf("--- Host: %s, version: %s ---\n", stat.Host, stat.Version)
	}
	printStatsDetails(allDocs, span)
	printGlobalLockDetails(allDocs, span)
	printLatencyDetails(allDocs, span)
	printMetricsDetails(allDocs, span)
	printWiredTigerCacheDetails(allDocs, span)
	printWiredTigerConcurrentTransactionsDetails(allDocs, span)
}

// printStatsDetails -
func printStatsDetails(docs []ServerStatusDoc, span int) {
	stat1 := ServerStatusDoc{}
	stat2 := ServerStatusDoc{}
	cnt := 0
	fmt.Println("\n--- Analytic Summary ---")
	fmt.Printf("+-------------------------+-------+-------+------+--------+--------+--------+--------+--------+--------+--------+\n")
	fmt.Printf("| Date/Time               | res   | virt  | fault| Command| Delete | Getmore| Insert | Query  | Update | iops   |\n")
	fmt.Printf("|-------------------------|-------+-------|------|--------|--------|--------|--------|--------|--------|--------|\n")
	for _, doc := range docs {
		buf, _ := json.Marshal(doc)
		json.Unmarshal(buf, &stat2)
		if cnt > 0 && stat2.Host == stat1.Host {
			d := int(stat2.LocalTime.Sub(stat1.LocalTime).Seconds())
			if cnt == len(docs)-1 || d >= span {
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
				stat1 = stat2
			}
		} else if cnt == 0 {
			stat1 = stat2
		}
		cnt++
	}
	fmt.Printf("+-------------------------+-------+-------+------+--------+--------+--------+--------+--------+--------+--------+\n")
}

// printLatencyDetails -
func printLatencyDetails(docs []ServerStatusDoc, span int) {
	stat1 := ServerStatusDoc{}
	stat2 := ServerStatusDoc{}
	cnt := 0
	fmt.Println("\n--- Latencies Summary (ms) ---")
	fmt.Printf("+-------------------------+----------+----------+----------+\n")
	fmt.Printf("| Date/Time               | reads    | writes   | commands |\n")
	fmt.Printf("|-------------------------|----------|----------|----------|\n")
	for _, doc := range docs {
		buf, _ := json.Marshal(doc)
		json.Unmarshal(buf, &stat2)
		if cnt > 0 && stat2.Host == stat1.Host {
			d := int(stat2.LocalTime.Sub(stat1.LocalTime).Seconds())
			if cnt == len(docs)-1 || d >= span {
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
				stat1 = stat2
			}
		} else if cnt == 0 {
			stat1 = stat2
		}
		cnt++
	}
	fmt.Printf("+-------------------------+----------+----------+----------+\n")
}

// printMetricsDetails -
func printMetricsDetails(docs []ServerStatusDoc, span int) {
	stat1 := ServerStatusDoc{}
	stat2 := ServerStatusDoc{}
	cnt := 0
	fmt.Println("\n--- Metrics ---")
	fmt.Printf("+-------------------------+----------+------------+------------+--------------+----------+----------+----------+----------+\n")
	fmt.Printf("| Date/Time               | Scanned  | ScannedObj |ScanAndOrder|WriteConflicts| Deleted  | Inserted | Returned | Updated  |\n")
	fmt.Printf("|-------------------------|----------|------------|------------|--------------|----------|----------|----------|----------|\n")
	for _, doc := range docs {
		buf, _ := json.Marshal(doc)
		json.Unmarshal(buf, &stat2)
		if cnt > 0 && stat2.Host == stat1.Host {
			d := int(stat2.LocalTime.Sub(stat1.LocalTime).Seconds())
			if cnt == len(docs)-1 || d >= span {
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
				stat1 = stat2
			}
		} else if cnt == 0 {
			stat1 = stat2
		}
		cnt++
	}
	fmt.Printf("+-------------------------+----------+------------+------------+--------------+----------+----------+----------+----------+\n")
}

// printGlobalLockDetails prints globalLock stats
func printGlobalLockDetails(docs []ServerStatusDoc, span int) {
	stat1 := ServerStatusDoc{}
	stat2 := ServerStatusDoc{}
	cnt := 0
	fmt.Println("\n--- Global Locks Summary ---")
	fmt.Printf("+-------------------------+--------------+--------------------------------------------+--------------------------------------------+\n")
	fmt.Printf("|                         | Total Time   | Active Clients                             | Current Queue                              |\n")
	fmt.Printf("| Date/Time               | (ms)         | total        | readers      | writers      | total        | readers      | writers      |\n")
	fmt.Printf("|-------------------------|--------------|--------------|--------------|--------------|--------------|--------------|--------------|\n")
	for _, doc := range docs {
		buf, _ := json.Marshal(doc)
		json.Unmarshal(buf, &stat2)
		if cnt > 0 && stat2.Host == stat1.Host {
			d := int(stat2.LocalTime.Sub(stat1.LocalTime).Seconds())
			if cnt == len(docs)-1 || d >= span {
				fmt.Printf("|%-25s|%14d|%14d|%14d|%14d|%14d|%14d|%14d|\n",
					stat2.LocalTime.In(loc).Format(time.RFC3339),
					(stat2.GlobalLock.TotalTime-stat1.GlobalLock.TotalTime)/1000,
					stat2.GlobalLock.CurrentQueue.Total,
					stat2.GlobalLock.CurrentQueue.Readers,
					stat2.GlobalLock.CurrentQueue.Writers,
					stat2.GlobalLock.CurrentQueue.Total,
					stat2.GlobalLock.CurrentQueue.Readers,
					stat2.GlobalLock.CurrentQueue.Writers)
				stat1 = stat2
			}
		} else if cnt == 0 {
			stat1 = stat2
		}
		cnt++
	}
	fmt.Printf("+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+--------------+\n")
}

// printWiredTigerCacheDetails prints wiredTiger cache stats
func printWiredTigerCacheDetails(docs []ServerStatusDoc, span int) {
	stat1 := ServerStatusDoc{}
	stat2 := ServerStatusDoc{}
	cnt := 0
	fmt.Println("\n--- WiredTiger Cache Summary ---")
	fmt.Printf("+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+--------------+\n")
	fmt.Printf("|                         | MaxBytes     | Currently    | Tracked      | Modified     | Unmodified   | PagesRead    | PagesWritten |\n")
	fmt.Printf("| Date/Time               | Configured   | InCache      | DirtyBytes   | PagesEvicted | PagesEvicted | IntoCache    | FromCache    |\n")
	fmt.Printf("|-------------------------|--------------|--------------|--------------|--------------|--------------|--------------|--------------|\n")
	for _, doc := range docs {
		buf, _ := json.Marshal(doc)
		json.Unmarshal(buf, &stat2)
		if cnt > 0 && stat2.Host == stat1.Host {
			d := int(stat2.LocalTime.Sub(stat1.LocalTime).Seconds())
			if cnt == len(docs)-1 || d >= span {
				fmt.Printf("|%-25s|%14d|%14d|%14d|%14d|%14d|%14d|%14d|\n",
					stat2.LocalTime.In(loc).Format(time.RFC3339),
					stat2.WiredTiger.Cache.MaxBytesConfigured,
					stat2.WiredTiger.Cache.CurrentlyInCache,
					stat2.WiredTiger.Cache.TrackedDirtyBytes,
					stat2.WiredTiger.Cache.ModifiedPagesEvicted,
					stat2.WiredTiger.Cache.UnmodifiedPagesEvicted,
					stat2.WiredTiger.Cache.PagesReadIntoCache-stat1.WiredTiger.Cache.PagesReadIntoCache,
					stat2.WiredTiger.Cache.PagesWrittenFromCache-stat1.WiredTiger.Cache.PagesWrittenFromCache)
				stat1 = stat2
			}
		} else if cnt == 0 {
			stat1 = stat2
		}
		cnt++
	}
	fmt.Printf("+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+--------------+\n")
}

// printWiredTigerConcurrentTransactionsDetails prints wiredTiger concurrentTransactions stats
func printWiredTigerConcurrentTransactionsDetails(docs []ServerStatusDoc, span int) {
	stat1 := ServerStatusDoc{}
	stat2 := ServerStatusDoc{}
	cnt := 0
	fmt.Println("\n--- WiredTiger Concurrent Transactions Summary ---")
	fmt.Printf("+-------------------------+--------------------------------------------+--------------------------------------------+\n")
	fmt.Printf("|                         | Read Ticket                                | Write Ticket                               |\n")
	fmt.Printf("| Date/Time               | Available    | Out          | Total        | Available    | Out          | Total        |\n")
	fmt.Printf("|-------------------------|--------------|--------------|--------------|--------------|--------------|--------------|\n")
	for _, doc := range docs {
		buf, _ := json.Marshal(doc)
		json.Unmarshal(buf, &stat2)
		if cnt > 0 && stat2.Host == stat1.Host {
			d := int(stat2.LocalTime.Sub(stat1.LocalTime).Seconds())
			if cnt == len(docs)-1 || d >= span {
				fmt.Printf("|%-25s|%14d|%14d|%14d|%14d|%14d|%14d|\n",
					stat2.LocalTime.In(loc).Format(time.RFC3339),
					stat2.WiredTiger.ConcurrentTransactions.Read.Available,
					stat2.WiredTiger.ConcurrentTransactions.Read.Out,
					stat2.WiredTiger.ConcurrentTransactions.Read.TotalTickets,
					stat2.WiredTiger.ConcurrentTransactions.Write.Available,
					stat2.WiredTiger.ConcurrentTransactions.Write.Out,
					stat2.WiredTiger.ConcurrentTransactions.Write.TotalTickets)
				stat1 = stat2
			}
		} else if cnt == 0 {
			stat1 = stat2
		}
		cnt++
	}
	fmt.Printf("+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+\n")
}

func getKeyFromReplicaSetName(uri string) string {
	dialInfo, _ := mgo.ParseURL(uri)
	key := dialInfo.ReplicaSetName
	if dialInfo.ReplicaSetName == "" {
		return "standalone"
	}
	return key
}
