// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/simagix/keyhole/mongo"
)

// ServerStatusDoc server status document
type ServerStatusDoc mongo.ServerStatusDoc

// ReplSetStatusDoc replset document
type ReplSetStatusDoc mongo.ReplSetStatusDoc

var keyholeStatsDataFile = os.TempDir() + "/keyhole_stats." + strings.Replace(time.Now().Format(time.RFC3339)[:19], ":", "", -1)
var loc, _ = time.LoadLocation("Local")
var mb = 1024.0 * 1024
var serverStatusDocs = map[string][]bson.M{}
var replSetStatusDocs = map[string][]ReplSetStatusDoc{}

// CollectServerStatus collects db.serverStatus() every minute
func (b Base) CollectServerStatus(dialInfo *mgo.DialInfo, uri string, channel chan string) {
	var pstat = ServerStatusDoc{}
	var stat = ServerStatusDoc{}
	var iop int
	var piop int
	var wSeconds = 10
	var r, w, c float64
	if b.verbose {
		rstr := fmt.Sprintf("CollectServerStatus collects every %d seconds(s)\n", wSeconds)
		channel <- rstr
	}

	mapKey := dialInfo.ReplicaSetName
	if mapKey == "" {
		mapKey = mongo.STANDALONE
	}
	channel <- "[" + mapKey + "] CollectServerStatus begins\n"
	for {
		session, err := mgo.DialWithInfo(dialInfo)
		if err == nil {
			serverStatus, _ := mongo.RunAdminCommand(session, "serverStatus")
			buf, _ := json.Marshal(serverStatus)
			json.Unmarshal(buf, &stat)
			serverStatusDocs[uri] = append(serverStatusDocs[uri], serverStatus)
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
func (b Base) ReplSetGetStatus(dialInfo *mgo.DialInfo, uri string, channel chan string) {
	var replSetStatus = ReplSetStatusDoc{}
	var doc bson.M
	var mapKey = dialInfo.ReplicaSetName
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
		session, err := mgo.DialWithInfo(dialInfo)
		if err == nil {
			doc, err = mongo.RunAdminCommand(session, "replSetGetStatus")
			if err == nil {
				buf, _ := json.Marshal(doc)
				json.Unmarshal(buf, &replSetStatus)
				replSetStatusDocs[uri] = append(replSetStatusDocs[uri], replSetStatus)

				if b.monitor == false {
					sort.Slice(replSetStatus.Members, func(i, j int) bool { return replSetStatus.Members[i].Name < replSetStatus.Members[j].Name })
					var ts int64
					for _, mb := range replSetStatus.Members {
						if mb.State == 1 {
							ts = GetOptime(mb.Optime)
							break
						}
					}

					str := fmt.Sprintf("[%s] replication lags: ", mapKey)
					for _, mb := range replSetStatus.Members {
						if mb.State == 2 {
							str += " - " + mb.Name + ": " + strconv.Itoa(int(ts-GetOptime(mb.Optime)))
						}
					}
					channel <- str
				}
			}
			session.Close()
		}
		time.Sleep(time.Duration(60) * time.Second)
	}
}

// CollectDBStats collects dbStats every 10 seconds
func (b Base) CollectDBStats(dialInfo *mgo.DialInfo, channel chan string, dbName string) {
	var docs map[string]interface{}
	var prevDataSize float64
	var dataSize float64
	var err error
	prevTime := time.Now()
	now := prevTime
	mapKey := dialInfo.ReplicaSetName
	if mapKey == "" {
		mapKey = mongo.STANDALONE
	}
	channel <- "[" + mapKey + "] CollectDBStats begins\n"
	session, err := mgo.DialWithInfo(dialInfo)
	defer session.Close()
	for i := 0; i < 10; i++ { // no need to collect after first 1.5 minutes
		if err == nil {
			stat, _ := mongo.RunCommandOnDB(session, "dbStats", dbName)
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
	var dialInfo *mgo.DialInfo
	if dialInfo, err = mongo.ParseURL(uri); err != nil {
		return filename, err
	}
	if err = mongo.AddCertificates(dialInfo, b.sslCAFile, b.sslPEMKeyFile); err != nil {
		return filename, err
	}
	if session, err = mgo.DialWithInfo(dialInfo); err != nil {
		return filename, err
	}
	defer session.Close()
	serverStatus, _ := mongo.RunAdminCommand(session, "serverStatus")
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
	dialInfo, _ := mongo.ParseURL(uri)
	mapKey := dialInfo.ReplicaSetName
	if mapKey == "" {
		mapKey = mongo.STANDALONE
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
