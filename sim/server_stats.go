// Copyright 2018 Kuei-chun Chen. All rights reserved.

package sim

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/simagix/gox"
	"github.com/simagix/keyhole/mdb"
	ftdc "github.com/simagix/mongo-ftdc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

var fileTimestamp = strings.Replace(time.Now().Format(time.RFC3339)[:19], ":", "", -1)
var keyholeStatsDataFile = "./keyhole_stats." + fileTimestamp

var mb = 1024.0 * 1024
var serverInfoDocs = map[string]ftdc.ServerInfoDoc{}
var serverStatusDocs = map[string][]ftdc.ServerStatusDoc{}
var replSetStatusDocs = map[string][]ftdc.ReplSetStatusDoc{}

func getServerInfoDocs(key string) ftdc.ServerInfoDoc         { return serverInfoDocs[key] }
func getServerStatusDocs(key string) []ftdc.ServerStatusDoc   { return serverStatusDocs[key] }
func getReplSetStatusDocs(key string) []ftdc.ReplSetStatusDoc { return replSetStatusDocs[key] }

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
	freq := Frequency{serverStatus: 5, replset: 10}
	return &ServerStats{channel: channel, uri: uri, mkey: getReplicaSetName(uri), freq: freq}
}

// SetVerbose sets verbose
func (st *ServerStats) SetVerbose(verbose bool) { st.verbose = verbose }

// SetPeekingMode sets peeking mode
func (st *ServerStats) SetPeekingMode(peek bool) {
	st.peek = peek
	if st.peek {
		st.freq.serverStatus = 1
		st.freq.replset = 5
	}
}

// getServerStatus gets db.serverStatus() every minute
func (st *ServerStats) getServerStatus(client *mongo.Client) error {
	var pstat = ftdc.ServerStatusDoc{}
	var stat = ftdc.ServerStatusDoc{}
	var iop int
	var piop int
	var r, w, c float64
	if st.verbose {
		rstr := fmt.Sprintf("getServerStatus gets every %d seconds(s)\n", st.freq.serverStatus)
		st.channel <- rstr
	}
	st.channel <- "[" + st.mkey + "] getServerStatus begins"
	for {
		serverStatus, _ := mdb.RunAdminCommand(client, "serverStatus")
		buf, _ := bson.Marshal(serverStatus)
		bson.Unmarshal(buf, &stat)
		serverStatusDocs[st.uri] = append(serverStatusDocs[st.uri], stat)
		if len(serverStatusDocs[st.uri]) > 600 {
			var err error
			var data []byte
			if data, err = getServerStatusData(st.uri); err != nil {
				return err
			}
			filename := keyholeStatsDataFile + "-" + st.mkey + ".gz"
			gox.OutputGzipped(data, filename)
			// also resets
			serverStatusDocs[st.uri] = []ftdc.ServerStatusDoc{}
			replSetStatusDocs[st.uri] = []ftdc.ReplSetStatusDoc{}
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
				str += fmt.Sprintf(", page faults: %d, iops: %.1f",
					(stat.ExtraInfo.PageFaults - pstat.ExtraInfo.PageFaults), iops)
				msg1 = fmt.Sprintf("[%s] CRUD+  - insert: %d, find: %d, update: %d, delete: %d, getmore: %d, command: %d",
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
				msg2 = fmt.Sprintf("[%s] Latency- read: %.1f, write: %.1f, command: %.1f (ms)",
					st.mkey, r, w, c)
			} else {
				str += ""
			}
		} else {
			str += ""
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
	var err error
	var replSetStatus = ftdc.ReplSetStatusDoc{}
	var doc bson.M
	if st.verbose {
		st.channel <- "ReplSetGetStatus gets every minute"
	}
	st.channel <- "[" + st.mkey + "] ReplSetGetStatus begins"

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
					ts = ftdc.GetOptime(mb.Optime)
					break
				}
			}

			str := fmt.Sprintf("[%s] replication lags: ", st.mkey)
			for _, mb := range replSetStatus.Members {
				if mb.State == 2 {
					str += " - " + mb.Name + ": " + strconv.Itoa(int(ts-ftdc.GetOptime(mb.Optime)))
				}
			}
			st.channel <- str
		}
		time.Sleep(time.Duration(st.freq.replset) * time.Second)
	}
}

// getDBStats gets dbStats every 10 seconds
func (st *ServerStats) getDBStats(client *mongo.Client, dbName string) error {
	var err error
	var docs map[string]interface{}
	var prevDataSize float64
	var dataSize float64
	prevTime := time.Now()
	now := prevTime
	st.channel <- "[" + st.mkey + "] getDBStats begins"
	for i := 0; i < 30; i++ { // no need to get after 5 minutes
		if err == nil {
			stat, _ := mdb.RunCommandOnDB(client, "dbStats", dbName)
			buf, _ := json.Marshal(stat)
			json.Unmarshal(buf, &docs)
			if docs["dataSize"] != nil {
				dataSize = mdb.ToFloat64(docs["dataSize"])
			}
			sec := now.Sub(prevTime).Seconds()
			delta := (dataSize - prevDataSize) / mb / sec
			if sec > 5 && delta >= 0 {
				str := fmt.Sprintf("[%s] Storage: %.1f -> %.1f, rate: %.1f MB/sec",
					st.mkey, prevDataSize/mb, dataSize/mb, delta)
				st.channel <- str
			}
			prevDataSize = dataSize
			prevTime = now
			now = time.Now()
		}
		time.Sleep(10 * time.Second)
	}
	st.channel <- "[" + st.mkey + "] getDBStats exiting..."
	return err
}

// getMongoConfig gets mongo configs
func (st *ServerStats) getMongoConfig(client *mongo.Client) error {
	var err error
	st.channel <- "[" + st.mkey + "] getMongoConfig begins"
	serverInfoDocs[st.uri] = ftdc.ServerInfoDoc{}
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
	st.channel <- "[" + st.mkey + "] mongo version: " + serverInfoDocs[st.uri].BuildInfo.Version
	return err
}

func (st *ServerStats) collectMetrics(client *mongo.Client, key string) {
	wtc := mdb.NewWiredTigerCache("server-stats")
	metrics := ftdc.NewMetrics()
	metrics.ProcessFiles([]string{})
	for {
		if err := wtc.GetAllDatabasesStats(client); err != nil {
			log.Println(err)
		}
		diag := ftdc.DiagnosticData{}
		diag.ServerInfo = getServerInfoDocs(key)
		diag.ServerStatusList = getServerStatusDocs(key)
		diag.ReplSetStatusList = getReplSetStatusDocs(key)
		metrics.AddFTDCDetailStats(&diag)
		time.Sleep(5 * time.Second)
	}
}

func getServerStatusData(uri string) ([]byte, error) {
	var err error
	var buf bytes.Buffer
	var data []byte
	if data, err = json.Marshal(serverStatusDocs[uri]); err != nil {
		return data, err
	}
	buf.Write(data)
	buf.Write([]byte{'\n'})
	if data, err = json.Marshal(replSetStatusDocs[uri]); err != nil {
		return data, err
	}
	buf.Write(data)
	buf.Write([]byte{'\n'})
	if data, err = json.Marshal(serverInfoDocs[uri]); err != nil {
		return data, err
	}
	buf.Write(data)
	buf.Write([]byte{'\n'})
	return buf.Bytes(), err
}

func getReplicaSetName(uri string) string {
	connStr, _ := connstring.Parse(uri)
	name := connStr.ReplicaSet
	if name == "" {
		name = mdb.Standalone
	}
	return name
}
