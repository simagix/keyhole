// Copyright 2018 Kuei-chun Chen. All rights reserved.

package sim

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

	"github.com/simagix/keyhole/mdb"
	anly "github.com/simagix/mongo-ftdc/analytics"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

var fileTimestamp = strings.Replace(time.Now().Format(time.RFC3339)[:19], ":", "", -1)
var keyholeStatsDataFile = "./keyhole_stats." + fileTimestamp

var mb = 1024.0 * 1024
var serverInfoDocs = map[string]anly.ServerInfoDoc{}
var serverStatusDocs = map[string][]anly.ServerStatusDoc{}
var replSetStatusDocs = map[string][]anly.ReplSetStatusDoc{}

func getServerInfoDocs(key string) anly.ServerInfoDoc         { return serverInfoDocs[key] }
func getServerStatusDocs(key string) []anly.ServerStatusDoc   { return serverStatusDocs[key] }
func getReplSetStatusDocs(key string) []anly.ReplSetStatusDoc { return replSetStatusDocs[key] }

// Frequency stores wait time
type Frequency struct {
	replset      int
	serverStatus int
}

// ServerStats stores server stats struct
type ServerStats struct {
	channel chan string
	mkey    string
	peek    bool
	freq    Frequency
	uri     string
	verbose bool
}

// NewServerStats gets server status
func NewServerStats(uri string, channel chan string) *ServerStats {
	connStr, _ := connstring.Parse(uri)
	mkey := connStr.ReplicaSet
	if mkey == "" {
		mkey = mdb.STANDALONE
	}
	freq := Frequency{serverStatus: 5, replset: 10}
	return &ServerStats{channel: channel, uri: uri, mkey: mkey, freq: freq}
}

// SetVerbose sets verbose
func (st *ServerStats) SetVerbose(verbose bool) { st.verbose = verbose }

// SetPeekingMode sets peeking mode
func (st *ServerStats) SetPeekingMode(peek bool) {
	st.peek = peek
	if st.peek == true {
		st.freq.serverStatus = 1
		st.freq.replset = 5
	}
}

// getServerStatus gets db.serverStatus() every minute
func (st *ServerStats) getServerStatus(client *mongo.Client) error {
	defer func() {
		if r := recover(); r != nil {
		}
	}()
	var pstat = anly.ServerStatusDoc{}
	var stat = anly.ServerStatusDoc{}
	var iop int
	var piop int
	var r, w, c float64
	if st.verbose {
		rstr := fmt.Sprintf("getServerStatus gets every %d seconds(s)\n", st.freq.serverStatus)
		st.channel <- rstr
	}
	st.channel <- "[" + st.mkey + "] getServerStatus begins\n"
	for {
		serverStatus, _ := mdb.RunAdminCommand(client, "serverStatus")
		buf, _ := bson.Marshal(serverStatus)
		bson.Unmarshal(buf, &stat)
		serverStatusDocs[st.uri] = append(serverStatusDocs[st.uri], stat)
		if len(serverStatusDocs[st.uri]) > 600 {
			st.saveServerStatusDocsToFile(st.uri)
		}

		var msg1, msg2 string
		str := fmt.Sprintf("[%s] Memory - resident: %d, virtual: %d",
			st.mkey, stat.Mem.Resident, stat.Mem.Virtual)
		iop = stat.Metrics.Document.Inserted + stat.Metrics.Document.Returned +
			stat.Metrics.Document.Updated + stat.Metrics.Document.Deleted
		iops := float64(iop-piop) / 60
		if len(serverStatusDocs[st.uri]) > 6 && len(serverStatusDocs[st.uri])%6 == 1 {
			pstat = serverStatusDocs[st.uri][len(serverStatusDocs[st.uri])-7]
			if stat.Host == pstat.Host {
				str += fmt.Sprintf(", page faults: %d, iops: %.1f\n",
					(stat.ExtraInfo.PageFaults - pstat.ExtraInfo.PageFaults), iops)
				msg1 = fmt.Sprintf("[%s] CRUD+  - insert: %d, find: %d, update: %d, delete: %d, getmore: %d, command: %d\n",
					st.mkey, stat.OpCounters.Insert-pstat.OpCounters.Insert,
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
					st.mkey, r, w, c)
			} else {
				str += "\n"
			}
		} else {
			str += "\n"
		}
		if len(serverStatusDocs[st.uri])%6 == 1 {
			st.channel <- str
			if msg1 != "" {
				st.channel <- msg1
			}
			if msg2 != "" {
				st.channel <- msg2
			}
		}
		piop = iop
		time.Sleep(time.Duration(st.freq.serverStatus) * time.Second)
	}
}

// getReplSetGetStatus gets {replSetGetStatus: 1} every minute
func (st *ServerStats) getReplSetGetStatus(client *mongo.Client) error {
	defer func() {
		if r := recover(); r != nil {
		}
	}()
	var err error
	var replSetStatus = anly.ReplSetStatusDoc{}
	var doc bson.M
	if st.verbose {
		rstr := fmt.Sprintf("ReplSetGetStatus gets every minute\n")
		st.channel <- rstr
	}
	st.channel <- "[" + st.mkey + "] ReplSetGetStatus begins\n"

	for {
		doc, err = mdb.RunAdminCommand(client, "replSetGetStatus")
		if err == nil {
			buf, _ := bson.Marshal(doc)
			bson.Unmarshal(buf, &replSetStatus)
			replSetStatusDocs[st.uri] = append(replSetStatusDocs[st.uri], replSetStatus)
			sort.Slice(replSetStatus.Members, func(i, j int) bool { return replSetStatus.Members[i].Name < replSetStatus.Members[j].Name })
			var ts int64
			for _, mb := range replSetStatus.Members {
				if mb.State == 1 {
					ts = anly.GetOptime(mb.Optime)
					break
				}
			}

			str := fmt.Sprintf("[%s] replication lags: ", st.mkey)
			for _, mb := range replSetStatus.Members {
				if mb.State == 2 {
					str += " - " + mb.Name + ": " + strconv.Itoa(int(ts-anly.GetOptime(mb.Optime)))
				}
			}
			st.channel <- str
		}
		time.Sleep(time.Duration(st.freq.replset) * time.Second)
	}
}

// getDBStats gets dbStats every 10 seconds
func (st *ServerStats) getDBStats(client *mongo.Client, dbName string) error {
	defer func() {
		if r := recover(); r != nil {
		}
	}()
	var err error
	var docs map[string]interface{}
	var prevDataSize float64
	var dataSize float64
	prevTime := time.Now()
	now := prevTime
	st.channel <- "[" + st.mkey + "] getDBStats begins\n"
	for i := 0; i < 30; i++ { // no need to get after 5 minutes
		if err == nil {
			stat, _ := mdb.RunCommandOnDB(client, "dbStats", dbName)
			buf, _ := json.Marshal(stat)
			json.Unmarshal(buf, &docs)
			if docs["dataSize"] != nil {
				dataSize = docs["dataSize"].(float64)
			}
			sec := now.Sub(prevTime).Seconds()
			delta := (dataSize - prevDataSize) / mb / sec
			if sec > 5 && delta >= 0 {
				str := fmt.Sprintf("[%s] Storage: %.1f -> %.1f, rate: %.1f MB/sec\n",
					st.mkey, prevDataSize/mb, dataSize/mb, delta)
				st.channel <- str
			}
			prevDataSize = dataSize
			prevTime = now
			now = time.Now()
		}
		time.Sleep(10 * time.Second)
	}
	st.channel <- "[" + st.mkey + "] getDBStats exiting...\n"
	return err
}

// getMongoConfig gets mongo configs
func (st *ServerStats) getMongoConfig(client *mongo.Client) error {
	defer func() {
		if r := recover(); r != nil {
		}
	}()
	var err error
	st.channel <- "[" + st.mkey + "] getMongoConfig begins\n"
	serverInfoDocs[st.uri] = anly.ServerInfoDoc{}
	var config = bson.M{}
	// hostInfo
	var hostInfo bson.M
	if hostInfo, err = mdb.RunAdminCommand(client, "hostInfo"); err == nil {
		config["hostInfo"] = hostInfo
	}

	// buildInfo
	var buildInfo bson.M
	if buildInfo, err = mdb.RunAdminCommand(client, "buildInfo"); err == nil {
		config["buildInfo"] = buildInfo
	}
	buf, _ := bson.Marshal(config)
	cfg := serverInfoDocs[st.uri]
	bson.Unmarshal(buf, &cfg)
	serverInfoDocs[st.uri] = cfg
	st.channel <- "[" + st.mkey + "] mongo version: " + serverInfoDocs[st.uri].BuildInfo.Version + "\n"
	return err
}

// printServerStatus prints serverStatusDocs summary for the duration
func (st *ServerStats) printServerStatus(client *mongo.Client) (string, error) {
	var err error
	var stat anly.ServerStatusDoc
	var filename string
	var str string
	serverStatus, _ := mdb.RunAdminCommand(client, "serverStatus")
	buf, _ := bson.Marshal(serverStatus)
	bson.Unmarshal(buf, &stat)
	serverStatusDocs[st.uri] = append(serverStatusDocs[st.uri], stat)
	if filename, err = st.saveServerStatusDocsToFile(st.uri); err != nil {
		return filename, err
	}
	var filenames = []string{filename}
	d := anly.NewDiagnosticData()
	if str, err = d.PrintDiagnosticData(filenames); err != nil {
		return filename, err
	}
	fmt.Println(str)
	return filename, err
}

// saveServerStatusDocsToFile appends []ServerStatusDoc to a file
func (st *ServerStats) saveServerStatusDocsToFile(uri string) (string, error) {
	var file *os.File
	var err error
	var filename string
	filename = keyholeStatsDataFile + "-" + st.mkey + ".gz"
	sbuf, _ := json.Marshal(serverStatusDocs[uri])
	serverStatusDocs[uri] = []anly.ServerStatusDoc{}
	rbuf, _ := json.Marshal(replSetStatusDocs[uri])
	replSetStatusDocs[uri] = []anly.ReplSetStatusDoc{}
	cbuf, _ := json.Marshal(serverInfoDocs[uri])
	var zbuf bytes.Buffer
	gz := gzip.NewWriter(&zbuf)
	gz.Write(sbuf)
	gz.Write([]byte{'\n'})
	gz.Write(rbuf)
	gz.Write([]byte{'\n'})
	gz.Write(cbuf)
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

func (st *ServerStats) collectMetrics(client *mongo.Client, key string) {
	wtc := mdb.NewWiredTigerCache(client)
	metrics := anly.NewMetrics()
	metrics.ProcessFiles([]string{})
	for {
		if err := wtc.GetAllDatabasesInfo(); err != nil {
			log.Println(err)
		}
		diag := anly.DiagnosticData{}
		diag.ServerInfo = getServerInfoDocs(key)
		diag.ServerStatusList = getServerStatusDocs(key)
		diag.ReplSetStatusList = getReplSetStatusDocs(key)
		metrics.AddFTDCDetailStats(&diag)
		time.Sleep(5 * time.Second)
	}
}
