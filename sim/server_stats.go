// Copyright 2018 Kuei-chun Chen. All rights reserved.

package sim

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
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

type serverStatusDoc anly.ServerStatusDoc
type replSetStatusDoc anly.ReplSetStatusDoc

var keyholeStatsDataFile = os.TempDir() + "/keyhole_stats." + strings.Replace(time.Now().Format(time.RFC3339)[:19], ":", "", -1)
var mb = 1024.0 * 1024
var serverStatusDocs = map[string][]serverStatusDoc{}
var replSetStatusDocs = map[string][]replSetStatusDoc{}

// CollectServerStatus collects db.serverStatus() every minute
func (rn *Runner) CollectServerStatus(uri string, channel chan string) {
	var err error
	var client *mongo.Client
	var ctx = context.Background()
	var pstat = serverStatusDoc{}
	var stat = serverStatusDoc{}
	var iop int
	var piop int
	var r, w, c float64
	span := 10 // defailt collects every 10 Seconds
	if rn.peek == true {
		span = 2
	}
	if rn.verbose {
		rstr := fmt.Sprintf("CollectServerStatus collects every %d seconds(s)\n", span)
		channel <- rstr
	}

	connStr, _ := connstring.Parse(uri)
	mapKey := connStr.ReplicaSet
	if mapKey == "" {
		mapKey = mdb.STANDALONE
	}
	channel <- "[" + mapKey + "] CollectServerStatus begins\n"
	for {
		if client, err = mdb.NewMongoClient(uri, rn.sslCAFile, rn.sslPEMKeyFile); err != nil {
			panic(err)
		}
		if err == nil {
			serverStatus, _ := mdb.RunAdminCommand(client, "serverStatus")
			buf, _ := bson.Marshal(serverStatus)
			bson.Unmarshal(buf, &stat)
			serverStatusDocs[uri] = append(serverStatusDocs[uri], stat)
			if len(serverStatusDocs[uri]) > 12 {
				rn.saveServerStatusDocsToFile(uri)
			}

			var msg1, msg2 string
			str := fmt.Sprintf("[%s] Memory - resident: %d, virtual: %d",
				mapKey, stat.Mem.Resident, stat.Mem.Virtual)
			iop = stat.Metrics.Document.Inserted + stat.Metrics.Document.Returned +
				stat.Metrics.Document.Updated + stat.Metrics.Document.Deleted
			iops := float64(iop-piop) / 60
			if len(serverStatusDocs[uri]) > 6 && len(serverStatusDocs[uri])%6 == 1 {
				pstat = serverStatusDocs[uri][len(serverStatusDocs[uri])-7]
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
			if rn.monitor == false && len(serverStatusDocs[uri])%6 == 1 {
				channel <- str
				if msg1 != "" {
					channel <- msg1
				}
				if msg2 != "" {
					channel <- msg2
				}
			}
			piop = iop
		}
		time.Sleep(time.Duration(span) * time.Second)
		client.Disconnect(ctx)
	}
}

// ReplSetGetStatus collects {replSetGetStatus: 1} every minute
func (rn *Runner) ReplSetGetStatus(uri string, channel chan string) {
	var err error
	var client *mongo.Client
	var ctx = context.Background()
	var replSetStatus = replSetStatusDoc{}
	var doc bson.M
	span := 60 // defailt collects every 60 Seconds
	if rn.peek == true {
		span = 2
	}
	connStr, _ := connstring.Parse(uri)
	mapKey := connStr.ReplicaSet
	if mapKey == "" {
		rstr := fmt.Sprintf("Not a replica set, collector exiting\n")
		channel <- rstr
		return
	}
	if rn.verbose {
		rstr := fmt.Sprintf("ReplSetGetStatus collects every minute\n")
		channel <- rstr
	}

	channel <- "[" + mapKey + "] ReplSetGetStatus begins\n"

	for {
		if client, err = mdb.NewMongoClient(uri, rn.sslCAFile, rn.sslPEMKeyFile); err != nil {
			channel <- err.Error()
			break
		}
		doc, err = mdb.RunAdminCommand(client, "replSetGetStatus")
		if err == nil {
			buf, _ := bson.Marshal(doc)
			bson.Unmarshal(buf, &replSetStatus)
			replSetStatusDocs[uri] = append(replSetStatusDocs[uri], replSetStatus)

			if rn.monitor == false {
				sort.Slice(replSetStatus.Members, func(i, j int) bool { return replSetStatus.Members[i].Name < replSetStatus.Members[j].Name })
				var ts int64
				for _, mb := range replSetStatus.Members {
					if mb.State == 1 {
						ts = anly.GetOptime(mb.Optime)
						break
					}
				}

				str := fmt.Sprintf("[%s] replication lags: ", mapKey)
				for _, mb := range replSetStatus.Members {
					if mb.State == 2 {
						str += " - " + mb.Name + ": " + strconv.Itoa(int(ts-anly.GetOptime(mb.Optime)))
					}
				}
				channel <- str
			}
		}
		time.Sleep(time.Duration(span) * time.Second)
		client.Disconnect(ctx)
	}
}

// CollectDBStats collects dbStats every 10 seconds
func (rn *Runner) CollectDBStats(client *mongo.Client, channel chan string, dbName string, uri string) {
	var docs map[string]interface{}
	var prevDataSize float64
	var dataSize float64
	var err error
	prevTime := time.Now()
	now := prevTime
	connStr, _ := connstring.Parse(uri)
	mapKey := connStr.ReplicaSet
	if mapKey == "" {
		mapKey = mdb.STANDALONE
	}
	channel <- "[" + mapKey + "] CollectDBStats begins\n"
	for i := 0; i < 10; i++ { // no need to collect after first 1.5 minutes
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
func (rn *Runner) PrintServerStatus(uri string, span int) (string, error) {
	var err error
	var client *mongo.Client
	var ctx = context.Background()
	var stat serverStatusDoc
	var filename string
	var str string
	if client, err = mdb.NewMongoClient(uri, rn.sslCAFile, rn.sslPEMKeyFile); err != nil {
		panic(err)
	}
	defer client.Disconnect(ctx)
	serverStatus, _ := mdb.RunAdminCommand(client, "serverStatus")
	buf, _ := bson.Marshal(serverStatus)
	bson.Unmarshal(buf, &stat)
	serverStatusDocs[uri] = append(serverStatusDocs[uri], stat)
	if filename, err = rn.saveServerStatusDocsToFile(uri); err != nil {
		return filename, err
	}
	var filenames = []string{filename}
	d := anly.NewDiagnosticData(span)
	if str, err = d.PrintDiagnosticData(filenames); err != nil {
		return filename, err
	}
	fmt.Println(str)
	return filename, err
}

// saveServerStatusDocsToFile appends []ServerStatusDoc to a file
func (rn *Runner) saveServerStatusDocsToFile(uri string) (string, error) {
	var file *os.File
	var err error
	var filename string
	connStr, _ := connstring.Parse(uri)
	mapKey := connStr.ReplicaSet
	if mapKey == "" {
		mapKey = mdb.STANDALONE
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
