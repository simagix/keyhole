// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

var keyholeStatsDataFile = os.TempDir() + "/keyhole_stats." + strings.Replace(time.Now().Format(time.RFC3339)[:19], ":", "", -1)
var loc, _ = time.LoadLocation("Local")
var mb = 1024.0 * 1024
var serverStatusDocs = map[string][]bson.M{}
var replSetStatusDocs = map[string][]ReplSetStatusDoc{}

// SHARDED cluster
const SHARDED = "sharded"

// STANDALONE cluster
const STANDALONE = "standalone"

// DocumentDoc contains db.serverStatus().document
type DocumentDoc struct {
	Deleted  int `json:"deleted" bson:"deleted"`
	Inserted int `json:"inserted" bson:"inserted"`
	Returned int `json:"returned" bson:"returned"`
	Updated  int `json:"updated" bson:"updated"`
}

// ExtraInfoDoc contains db.serverStatus().extra_info
type ExtraInfoDoc struct {
	PageFaults int64 `json:"page_faults" bson:"page_faults"`
}

// GlobalLockSubDoc contains db.serverStatus().globalLockDoc.[activeClients|currentQueue]
type GlobalLockSubDoc struct {
	Readers int64 `json:"readers" bson:"readers"`
	Total   int64 `json:"total" bson:"total"`
	Writers int64 `json:"writers" bson:"writers"`
}

// GlobalLockDoc contains db.serverStatus().globalLockDoc
type GlobalLockDoc struct {
	ActiveClients GlobalLockSubDoc `json:"activeClients" bson:"activeClients"`
	CurrentQueue  GlobalLockSubDoc `json:"currentQueue" bson:"currentQueue"`
	TotalTime     int              `json:"totalTime" bson:"totalTime"`
}

// MemDoc containers db.serverStatus().mem
type MemDoc struct {
	Resident int64 `json:"resident" bson:"resident"`
	Virtual  int64 `json:"virtual" bson:"virtual"`
}

// MetricsDoc contains db.serverStatus().metrics
type MetricsDoc struct {
	Document      DocumentDoc      `json:"document" bson:"document"`
	QueryExecutor QueryExecutorDoc `json:"queryExecutor" bson:"queryExecutor"`
	Operation     OperationDoc     `json:"operation" bson:"operation"`
}

// OperationDoc contains db.serverStatus().operation
type OperationDoc struct {
	ScanAndOrder   int64 `json:"scanAndOrder" bson:"scanAndOrder"`
	WriteConflicts int64 `json:"writeConflicts" bson:"writeConflicts"`
}

// OpCountersDoc contains db.serverStatus().OpCounters
type OpCountersDoc struct {
	Command int64 `json:"command" bson:"command"`
	Delete  int64 `json:"delete" bson:"delete"`
	Getmore int64 `json:"getmore" bson:"getmore"`
	Insert  int64 `json:"insert" bson:"insert"`
	Query   int64 `json:"query" bson:"query"`
	Update  int64 `json:"update" bson:"update"`
}

// OpLatenciesDoc contains db.serverStatus().opLatencies
type OpLatenciesDoc struct {
	Commands OpLatenciesOpDoc `json:"commands" bson:"commands"`
	Reads    OpLatenciesOpDoc `json:"reads" bson:"reads"`
	Writes   OpLatenciesOpDoc `json:"writes" bson:"writes"`
}

// OpLatenciesOpDoc contains doc of db.serverStatus().opLatencies
type OpLatenciesOpDoc struct {
	Latency int64 `json:"latency" bson:"latency"`
	Ops     int64 `json:"ops" bson:"ops"`
}

// QueryExecutorDoc contains db.serverStatus().queryExecutor
type QueryExecutorDoc struct {
	Scanned        int64 `json:"scanned" bson:"scanned"`
	ScannedObjects int64 `json:"scannedObjects" bson:"scannedObjects"`
}

// WiredTigerCacheDoc contains db.serverStatus().wiredTiger.cache
type WiredTigerCacheDoc struct {
	MaxBytesConfigured     int64 `json:"maximum bytes configured" bson:"maximum bytes configured"`
	CurrentlyInCache       int64 `json:"bytes currently in the cache" bson:"bytes currently in the cache"`
	ModifiedPagesEvicted   int64 `json:"modified pages evicted" bson:"modified pages evicted"`
	UnmodifiedPagesEvicted int64 `json:"unmodified pages evicted" bson:"unmodified pages evicted"`
	TrackedDirtyBytes      int64 `json:"tracked dirty bytes in the cache" bson:"tracked dirty bytes in the cache"`
	PagesReadIntoCache     int64 `json:"pages read into cache" bson:"pages read into cache"`
	PagesWrittenFromCache  int64 `json:"pages written from cache" bson:"pages written from cache"`
}

// ConcurrentTransactionsCountDoc contains db.serverStatus().wiredTiger.concurrentTransactions.[read|write]
type ConcurrentTransactionsCountDoc struct {
	Available    int64 `json:"available" bson:"available"`
	Out          int64 `json:"out" bson:"out"`
	TotalTickets int64 `json:"totalTickets" bson:"totalTickets"`
}

// ConcurrentTransactionsDoc contains db.serverStatus().wiredTiger.concurrentTransactions
type ConcurrentTransactionsDoc struct {
	Read  ConcurrentTransactionsCountDoc `json:"read" bson:"read"`
	Write ConcurrentTransactionsCountDoc `json:"write" bson:"write"`
}

// WiredTigerDoc contains db.serverStatus().wiredTiger
type WiredTigerDoc struct {
	Perf                   interface{}               `json:"perf" bson:"perf"`
	Cache                  WiredTigerCacheDoc        `json:"cache" bson:"cache"`
	ConcurrentTransactions ConcurrentTransactionsDoc `json:"concurrentTransactions" bson:"concurrentTransactions"`
}

// ConnectionsDoc contains db.serverStatus().connections
type ConnectionsDoc struct {
	Current      int64 `json:"current" bson:"current"`
	Available    int64 `json:"available" bson:"available"`
	TotalCreated int64 `json:"totalCreated" bson:"totalCreated"`
}

// ServerStatusDoc contains docs from db.serverStatus()
type ServerStatusDoc struct {
	Connections ConnectionsDoc `json:"connections" bson:"connections"`
	ExtraInfo   ExtraInfoDoc   `json:"extra_info" bson:"extra_info"`
	GlobalLock  GlobalLockDoc  `json:"globalLock" bson:"globalLock"`
	Host        string         `json:"host" bson:"host"`
	LocalTime   time.Time      `json:"localTime" bson:"localTime"`
	Mem         MemDoc         `json:"mem" bson:"mem"`
	Metrics     MetricsDoc     `json:"metrics" bson:"metrics"`
	OpCounters  OpCountersDoc  `json:"opcounters" bson:"opcounters"`
	OpLatencies OpLatenciesDoc `json:"opLatencies" bson:"opLatencies"`
	Process     string         `json:"process" bson:"process"`
	Repl        interface{}    `json:"repl" bson:"repl"`
	Sharding    interface{}    `json:"sharding" bson:"sharding"`
	Uptime      int64          `json:"uptime" bson:"uptime"`
	Version     string         `json:"version" bson:"version"`
	WiredTiger  WiredTigerDoc  `json:"wiredTiger" bson:"wiredTiger"`
}

// CollectServerStatus collects db.serverStatus() every minute
func (b Base) CollectServerStatus(uri string, channel chan string) {
	pstat := ServerStatusDoc{}
	stat := ServerStatusDoc{}
	var iop int
	var piop int
	wSeconds := 10
	var r, w, c float64
	if b.verbose {
		rstr := fmt.Sprintf("CollectServerStatus collects every %d seconds(s)\n", wSeconds)
		channel <- rstr
	}

	var dialInfo *mgo.DialInfo
	var err error
	if dialInfo, err = ParseDialInfo(uri); err != nil {
		return
	}

	mapKey := dialInfo.ReplicaSetName
	if mapKey == "" {
		mapKey = STANDALONE
	}
	channel <- "[" + mapKey + "] CollectServerStatus begins\n"
	for {
		session, err := GetSession(dialInfo, false, b.ssl, b.sslCAFile, b.sslPEMKeyFile)
		if err == nil {
			serverStatus := bson.M{}
			session.DB("admin").Run("serverStatus", &serverStatus)
			buf, _ := json.Marshal(serverStatus)
			json.Unmarshal(buf, &stat)
			serverStatusDocs[uri] = append(serverStatusDocs[uri], serverStatus)
			// dkey := dialInfo.ReplicaSetName + "/" + strings.Join(dialInfo.Addrs[:], ",")
			// ChartsDocs[dkey] = append(ChartsDocs[dkey], serverStatus)
			// for len(ChartsDocs[dkey]) > 60 { // shift
			// 	ChartsDocs[dkey] = ChartsDocs[dkey][1:]
			// }
			if len(serverStatusDocs[uri]) > 12 {
				b.saveServerStatusDocsToFile(uri)
			}

			var msg1, msg2 string
			str := fmt.Sprintf("[%s] Memory - resident: %d, virtual: %d",
				mapKey, stat.Mem.Resident, stat.Mem.Virtual)
			iop = stat.Metrics.Document.Inserted + stat.Metrics.Document.Returned +
				stat.Metrics.Document.Updated + stat.Metrics.Document.Deleted
			iops := float64(iop-piop) / 60
			if len(serverStatusDocs[uri]) > 6 && len(serverStatusDocs[uri])%6 == 1 {
				buf, _ = json.Marshal(serverStatusDocs[uri][len(serverStatusDocs[uri])-7])
				json.Unmarshal(buf, &pstat)
				if stat.Host == pstat.Host {
					str += fmt.Sprintf(", page faults: %d, iops: %.1f\n",
						(stat.ExtraInfo.PageFaults - pstat.ExtraInfo.PageFaults), iops)
					msg1 = fmt.Sprintf("[%s] CRUD+  - insert: %d, find: %d, update: %d, delete: %d, getmore: %d, command: %d\n",
						mapKey, stat.OpCounters.Insert-pstat.OpCounters.Insert,
						stat.OpCounters.Query-pstat.OpCounters.Query,
						stat.OpCounters.Update-pstat.OpCounters.Update,
						stat.OpCounters.Delete-pstat.OpCounters.Delete,
						stat.OpCounters.Getmore-pstat.OpCounters.Getmore,
						stat.OpCounters.Command-pstat.OpCounters.Command)
					r = 0
					if stat.OpLatencies.Reads.Ops > 0 {
						r = float64(stat.OpLatencies.Reads.Latency) / float64(stat.OpLatencies.Reads.Ops) / 1000
					}
					w = 0
					if stat.OpLatencies.Writes.Ops > 0 {
						w = float64(stat.OpLatencies.Writes.Latency) / float64(stat.OpLatencies.Writes.Ops) / 1000
					}
					c = 0
					if stat.OpLatencies.Commands.Ops > 0 {
						c = float64(stat.OpLatencies.Commands.Latency) / float64(stat.OpLatencies.Commands.Ops) / 1000
					}
					msg2 = fmt.Sprintf("[%s] Latency- read: %.1f, write: %.1f, command: %.1f (ms)\n",
						mapKey, r, w, c)
				} else {
					str += "\n"
				}
			} else {
				str += "\n"
			}
			if b.monitor == false && len(serverStatusDocs[uri])%6 == 1 {
				channel <- str
				if msg1 != "" {
					channel <- msg1
				}
				if msg2 != "" {
					channel <- msg2
				}
			}
			piop = iop
			session.Close()
		}
		time.Sleep(time.Duration(wSeconds) * time.Second)
	}
}

// ReplSetGetStatus collects {replSetGetStatus: 1} every minute
func (b Base) ReplSetGetStatus(uri string, channel chan string) {
	var replSetStatus = ReplSetStatusDoc{}
	var doc bson.M

	var dialInfo *mgo.DialInfo
	var err error
	if dialInfo, err = ParseDialInfo(uri); err != nil {
		return
	}

	mapKey := dialInfo.ReplicaSetName
	if mapKey == "" {
		rstr := fmt.Sprintf("Not a replica set, collector exiting\n")
		channel <- rstr
		return
	}
	if b.verbose {
		rstr := fmt.Sprintf("ReplSetGetStatus collects every minute\n")
		channel <- rstr
	}

	channel <- "[" + mapKey + "] ReplSetGetStatus begins\n"

	for {
		session, err := GetSession(dialInfo, false, b.ssl, b.sslCAFile, b.sslPEMKeyFile)
		if err == nil {
			if doc, err = AdminCommand(session, "replSetGetStatus"); err != nil {
				log.Println(err)
				continue
			}

			buf, _ := json.Marshal(doc)
			json.Unmarshal(buf, &replSetStatus)
			replSetStatusDocs[uri] = append(replSetStatusDocs[uri], replSetStatus)

			if b.monitor == false {
				sort.Slice(replSetStatus.Members, func(i, j int) bool { return replSetStatus.Members[i].Name < replSetStatus.Members[j].Name })
				var ts int64
				for _, mb := range replSetStatus.Members {
					if mb.StateStr == PRIMARY {
						ts = GetOptime(mb.Optime)
						break
					}
				}

				str := fmt.Sprintf("[%s] replication lags: ", mapKey)
				for _, mb := range replSetStatus.Members {
					if mb.StateStr == SECONDARY {
						str += " - " + mb.Name + ": " + strconv.Itoa(int(ts-GetOptime(mb.Optime)))
					}
				}
				channel <- str
			}
			session.Close()
		}
		time.Sleep(time.Duration(60) * time.Second)
	}
}

// CollectDBStats collects dbStats every 10 seconds
func (b Base) CollectDBStats(uri string, channel chan string, dbName string) {
	var docs map[string]interface{}
	var prevDataSize float64
	var dataSize float64
	prevTime := time.Now()
	now := prevTime
	dialInfo, _ := ParseDialInfo(uri)
	mapKey := dialInfo.ReplicaSetName
	if mapKey == "" {
		mapKey = STANDALONE
	}
	channel <- "[" + mapKey + "] CollectDBStats begins\n"
	session, err := GetSession(dialInfo, false, b.ssl, b.sslCAFile, b.sslPEMKeyFile)
	defer session.Close()
	for i := 0; i < 10; i++ { // no need to collect after first 1.5 minutes
		if err == nil {
			stat := bson.M{}
			session.DB(dbName).Run("dbStats", &stat)
			buf, _ := json.Marshal(stat)
			json.Unmarshal(buf, &docs)
			if docs["dataSize"] != nil {
				dataSize = docs["dataSize"].(float64)
			}
			sec := now.Sub(prevTime).Seconds()
			delta := (dataSize - prevDataSize) / mb / sec
			if sec > 5 && delta >= 0 {
				str := fmt.Sprintf("[%s] Storage: %.1f -> %.1f, rate: %.1f MB/sec\n",
					mapKey, prevDataSize/mb, dataSize/mb, delta)
				channel <- str
			}
			prevDataSize = dataSize
			prevTime = now
			now = time.Now()
		}
		time.Sleep(10 * time.Second)
	}
	channel <- "[" + mapKey + "] CollectDBStats exiting...\n"
}

// PrintServerStatus prints serverStatusDocs summary for the duration
func (b Base) PrintServerStatus(uri string, span int) (string, error) {
	var session *mgo.Session
	var err error
	var filename string
	var str string

	dialInfo, _ := ParseDialInfo(uri)
	if session, err = GetSession(dialInfo, false, b.ssl, b.sslCAFile, b.sslPEMKeyFile); err != nil {
		return filename, err
	}
	defer session.Close()
	serverStatus := bson.M{}
	session.DB("admin").Run("serverStatus", &serverStatus)
	buf, _ := json.Marshal(serverStatus)
	json.Unmarshal(buf, &serverStatus)
	serverStatusDocs[uri] = append(serverStatusDocs[uri], serverStatus)
	if filename, err = b.saveServerStatusDocsToFile(uri); err != nil {
		return filename, err
	}
	d := NewDiagnosticData(span)
	var filenames = []string{filename}
	if str, err = d.PrintDiagnosticData(filenames, false); err != nil {
		return filename, err
	}
	fmt.Println(str)
	return filename, err
}

// saveServerStatusDocsToFile appends []ServerStatusDoc to a file
func (b Base) saveServerStatusDocsToFile(uri string) (string, error) {
	var file *os.File
	var err error
	var filename string
	dialInfo, _ := ParseDialInfo(uri)
	mapKey := dialInfo.ReplicaSetName
	if mapKey == "" {
		mapKey = STANDALONE
	}
	sbuf, _ := json.Marshal(serverStatusDocs[uri])
	serverStatusDocs[uri] = serverStatusDocs[uri][:0]
	filename = keyholeStatsDataFile + "-" + mapKey + ".gz"
	rbuf, _ := json.Marshal(replSetStatusDocs[uri])
	replSetStatusDocs[uri] = replSetStatusDocs[uri][:0]
	var zbuf bytes.Buffer
	gz := gzip.NewWriter(&zbuf)
	gz.Write(sbuf)
	gz.Write([]byte{'\n'})
	gz.Write(rbuf)
	gz.Write([]byte{'\n'})
	gz.Close() // close this before flushing the bytes to the buffer.

	if file, err = os.OpenFile(filename, os.O_WRONLY|os.O_APPEND, 0644); err != nil {
		if file, err = os.Create(filename); err != nil {
			return filename, err
		}
	}
	defer file.Close()
	file.Write(zbuf.Bytes())
	file.Sync()
	return filename, err
}
