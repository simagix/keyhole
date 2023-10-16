// Copyright 2018 Kuei-chun Chen. All rights reserved.

package sim

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/simagix/gox"
	"github.com/simagix/keyhole/mdb"
	ftdc "github.com/simagix/mongo-ftdc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

const (
	outdir = "./out"
)

// Runner -
type Runner struct {
	Logger  *gox.Logger         `bson:"keyhole"`
	Metrics map[string][]bson.M `bson:"metrics"`
	Results []string            `bson:"results"`

	auto           bool
	channel        chan string
	client         *mongo.Client
	clusterType    string
	collectionName string
	conns          int
	connString     connstring.ConnString
	dbName         string
	drop           bool
	duration       int
	filename       string
	mutex          sync.RWMutex
	peek           bool
	simOnly        bool
	tps            int
	txFilename     string
	uri            string
	uriList        []string
	verbose        bool
}

// NewRunner - Constructor
func NewRunner(connString connstring.ConnString) (*Runner, error) {
	var err error
	runner := Runner{Logger: gox.GetLogger("keyhole"), connString: connString, conns: runtime.NumCPU(),
		channel: make(chan string), collectionName: mdb.ExamplesCollection, Metrics: map[string][]bson.M{},
		mutex: sync.RWMutex{}}
	runner.dbName = connString.Database
	if runner.dbName == "" {
		runner.dbName = mdb.KeyholeDB
	}
	if runner.client, err = mdb.NewMongoClient(connString.String()); err != nil {
		return &runner, err
	}
	stats := mdb.NewClusterStats("")
	stats.GetClusterStatsSummary(runner.client)
	runner.clusterType = stats.Cluster
	if runner.clusterType == "" {
		runner.Logger.Warn("unable to retrieve cluster type")
	}
	runner.uriList = []string{connString.String()}
	if runner.clusterType == mdb.Sharded {
		if shards, err := mdb.GetShards(runner.client); err != nil {
			return &runner, err
		} else if runner.uriList, err = mdb.GetAllShardURIs(shards, connString); err != nil {
			return &runner, err
		}
	}
	runner.uri = runner.uriList[len(runner.uriList)-1]
	return &runner, nil
}

// SetCollection set collection name
func (rn *Runner) SetCollection(collectionName string) {
	if collectionName != "" {
		rn.collectionName = collectionName
	} else {
		rn.collectionName = mdb.ExamplesCollection
	}
}

// SetTPS set transaction per second
func (rn *Runner) SetTPS(tps int) {
	rn.tps = tps
}

// SetAutoMode set transaction per second
func (rn *Runner) SetAutoMode(auto bool) { rn.auto = auto }

// SetTemplateFilename -
func (rn *Runner) SetTemplateFilename(filename string) {
	rn.filename = filename
}

// SetVerbose sets verbose mode
func (rn *Runner) SetVerbose(verbose bool) {
	rn.verbose = verbose
}

// SetPeekingMode -
func (rn *Runner) SetPeekingMode(mode bool) {
	rn.peek = mode
	if rn.peek {
		go func(x int) {
			time.Sleep(time.Duration(x) * time.Minute)
			rn.terminate()
		}(rn.duration)
	}
}

// SetDuration sets simulation/load test duration
func (rn *Runner) SetDuration(duration int) {
	rn.duration = duration
}

// SetDropFirstMode -
func (rn *Runner) SetDropFirstMode(mode bool) {
	rn.drop = mode
}

// SetNumberConnections -
func (rn *Runner) SetNumberConnections(num int) {
	if num == 0 {
		return
	}
	rn.conns = num
}

// SetTransactionTemplate sets transaction template file
func (rn *Runner) SetTransactionTemplate(filename string) {
	rn.txFilename = filename
}

// SetSimOnlyMode -
func (rn *Runner) SetSimOnlyMode(mode bool) {
	rn.simOnly = mode
}

// Start process requests
func (rn *Runner) Start() error {
	var err error
	if rn.peek {
		return nil
	}
	if !rn.auto {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Begin a load test [y/N]: ")
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		if text != "y" && text != "Y" {
			os.Exit(0)
		}
	}
	rn.Logger.Info("Duration in minute(s):", rn.duration)
	if rn.dbName == "" || rn.dbName == "admin" || rn.dbName == "config" || rn.dbName == "local" {
		rn.dbName = mdb.KeyholeDB // switch to _KEYHOLE_88800 database for load tests
	}
	if rn.drop {
		rn.Cleanup()
	}
	rn.initSimDocs()
	tdoc := GetTransactions(rn.txFilename)
	// Simulation mode
	// 1st minute - build up data and memory
	// 2nd and 3rd minutes - normal TPS ops
	// remaining minutes - burst with no delay
	// last minute - normal TPS ops until exit
	rn.Logger.Info(fmt.Sprintf("Total TPS: %d (%d tps/conn * %d conns), duration: %d (mins)", rn.tps*rn.conns, rn.tps, rn.conns, rn.duration))
	simTime := rn.duration
	if !rn.simOnly {
		simTime--
		rn.createIndexes(tdoc.Indexes)
	}
	for i := 0; i < rn.conns; i++ {
		go func(thread int) {
			if !rn.simOnly && rn.duration > 0 {
				if err = rn.PopulateData(); err != nil {
					rn.Logger.Info("Thread", thread, "existing with", err)
					return
				}
				time.Sleep(10 * time.Millisecond)
			}

			if err = rn.Simulate(simTime, tdoc.Transactions, thread); err != nil {
				rn.Logger.Info("Thread", thread, "existing with", err)
				return
			}
		}(i)
	}
	return nil
}

func (rn *Runner) terminate() {
	var err error
	var client *mongo.Client
	var filename string
	var filenames []string
	var result string

	os.Mkdir(outdir, 0755)
	rn.Cleanup()
	rn.Results = []string{}
	for _, uri := range rn.uriList {
		if client, err = mdb.NewMongoClient(uri); err != nil {
			rn.Logger.Info(err)
			continue
		}

		var err error
		var stat ftdc.ServerStatusDoc
		serverStatus, _ := mdb.RunAdminCommand(client, "serverStatus")
		buf, _ := bson.Marshal(serverStatus)
		bson.Unmarshal(buf, &stat)
		serverStatusDocs[uri] = append(serverStatusDocs[uri], stat)

		var data []byte
		if data, err = getServerStatusData(uri); err != nil {
			rn.Logger.Error(err)
			break
		}

		// save metrics to a file
		filename := fmt.Sprintf("%v/%v-%v.gz", outdir, keyholeStatsDataFile, getReplicaSetName(uri))
		filenames = append(filenames, filename)
		gox.OutputGzipped(data, filename)
		d := ftdc.NewDiagnosticData()
		reader := bufio.NewReader(strings.NewReader(string(data)))
		if err = d.AnalyzeServerStatus(reader); err != nil {
			rn.Logger.Error(err)
			break
		}
		strs := []string{}
		if d.ServerInfo != nil {
			var p mdb.ClusterStats
			data, _ := json.Marshal(d.ServerInfo)
			json.Unmarshal(data, &p)
			result := fmt.Sprintf("\n* MongoDB v%v %v (%v) %v %v %v cores %v mem",
				p.BuildInfo.Version, p.HostInfo.System.Hostname, p.HostInfo.OS.Name,
				p.ServerStatus.Process, p.Cluster, p.HostInfo.System.NumCores, p.HostInfo.System.MemSizeMB)
			strs = append(strs, result)
		}
		strs = append(strs, ftdc.PrintAllStats(d.ServerStatusList, -1))
		result = strings.Join(strs, "\n")
		fmt.Println(result)
		rn.Results = append(rn.Results, result)
	}
	hostname, _ := os.Hostname()
	filename = fmt.Sprintf(`%s/%s.%s-perf.bson.gz`, outdir, hostname, fileTimestamp)
	var buf []byte
	if buf, err = bson.Marshal(rn); err != nil {
		rn.Logger.Info("marshal error:", err)
	}
	gox.OutputGzipped(buf, filename)
	filenames = append(filenames, filename)
	zipFile := fmt.Sprintf(`%s/%s.%s-perf.zip`, outdir, hostname, fileTimestamp)
	gox.ZipFiles(zipFile, filenames)
	rn.Logger.Info("stats written to ", zipFile)
	for _, f := range filenames {
		os.Remove(f)
	}
	os.Exit(0)
}

// CollectAllStatus collects all server stats
func (rn *Runner) CollectAllStatus() error {
	var err error
	for i, uri := range rn.uriList {
		var client *mongo.Client
		if client, err = mdb.NewMongoClient(uri); err != nil {
			rn.Logger.Info(err)
			continue
		}
		stats := NewServerStats(uri, rn.channel)
		stats.SetVerbose(rn.verbose)
		stats.SetPeekingMode(rn.peek)
		go stats.getDBStats(client, rn.dbName)
		go stats.getReplSetGetStatus(client)
		go stats.getServerStatus(client)
		go stats.getMongoConfig(client)
		if i == 0 {
			go stats.collectMetrics(client, uri)
		}
	}

	quit := make(chan os.Signal, 2)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	timer := time.NewTimer(time.Duration(rn.duration) * time.Minute)

	for {
		select {
		case <-quit:
			rn.terminate()
		case <-timer.C:
			rn.terminate()
		default:
			rn.Logger.Info(<-rn.channel)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// CreateIndexes creates indexes
func (rn *Runner) createIndexes(docs []bson.M) error {
	var err error
	var ctx = context.Background()
	c := rn.client.Database(rn.dbName).Collection(rn.collectionName)
	indexView := c.Indexes()
	idx := mongo.IndexModel{Keys: bson.D{{Key: "_search", Value: 1}}}
	if _, err = indexView.CreateOne(ctx, idx); err != nil {
		return err
	}
	if len(docs) == 0 {
		idx = mongo.IndexModel{Keys: bson.D{{Key: "email", Value: 1}}}
		if _, err = indexView.CreateOne(ctx, idx); err != nil {
			return err
		}

		if rn.clusterType == mdb.Sharded {
			if err = rn.splitChunks(); err != nil {
				fmt.Println(err)
			}
		}
	}

	for _, doc := range docs {
		keys := bson.D{}
		for k, v := range doc {
			x := int32(1)
			switch v := v.(type) {
			case int:
				if v < 0 {
					x = -1
				}
			case float64:
				if v < 0 {
					x = -1
				}
			}

			keys = append(keys, bson.E{Key: k, Value: x})
		}
		idx := mongo.IndexModel{
			Keys: keys,
		}
		if _, err = indexView.CreateOne(ctx, idx); err != nil {
			return err
		}
	}

	return err
}

// Cleanup drops the temp database
func (rn *Runner) Cleanup() error {
	var err error
	if rn.peek {
		return err
	}
	if !rn.simOnly && rn.dbName == mdb.KeyholeDB {
		ctx := context.Background()
		if rn.collectionName == mdb.ExamplesCollection {
			rn.Logger.Info("dropping collection", mdb.KeyholeDB, mdb.ExamplesCollection)
			if err = rn.client.Database(mdb.KeyholeDB).Collection(mdb.ExamplesCollection).Drop(ctx); err != nil {
				rn.Logger.Info(err)
			}
		}
		rn.Logger.Info("dropping temp database", mdb.KeyholeDB)
		if err = rn.client.Database(rn.dbName).Drop(ctx); err != nil {
			rn.Logger.Info(err)
		}
	}

	time.Sleep(time.Second)
	return err
}

func (rn *Runner) splitChunks() error {
	var err error
	var ctx = context.Background()
	var cursor *mongo.Cursor
	ns := rn.dbName + "." + rn.collectionName
	result := bson.M{}
	filter := bson.M{"_id": rn.dbName}
	if err = rn.client.Database("config").Collection("databases").FindOne(ctx, filter).Decode(&result); err != nil {
		return err
	}
	primary := result["primary"].(string)
	rn.Logger.Info("Sharding collection:", ns)
	cmd := bson.D{{Key: "enableSharding", Value: rn.dbName}}
	if err = rn.client.Database("admin").RunCommand(ctx, cmd).Decode(&result); err != nil {
		return err
	}
	cmd = bson.D{{Key: "shardCollection", Value: ns}, {Key: "key", Value: bson.M{"email": 1}}}
	if err = rn.client.Database("admin").RunCommand(ctx, cmd).Decode(&result); err != nil {
		return err
	}
	rn.Logger.Info("splitting chunks...")
	if cursor, err = rn.client.Database("config").Collection("shards").Find(ctx, bson.D{{}}); err != nil {
		return err
	}
	otherShards := []bson.M{}
	for cursor.Next(ctx) {
		v := bson.M{}
		if err = cursor.Decode(&v); err != nil {
			rn.Logger.Info(err)
			continue
		}
		if primary != v["_id"].(string) {
			otherShards = append(otherShards, v)
		}
	}
	shardKeys := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M",
		"N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
	divider := 1 + len(shardKeys)/(len(otherShards)+1)
	for i := range otherShards {
		cmd := bson.D{{Key: "split", Value: ns}, {Key: "middle", Value: bson.M{"email": shardKeys[(i+1)*divider]}}}
		if err = rn.client.Database("admin").RunCommand(ctx, cmd).Decode(&result); err != nil { // could be split already
			return err
		}
	}

	if len(otherShards) < 1 {
		return nil
	}
	rn.Logger.Info("moving chunks...")
	filter = bson.M{"ns": ns}
	opts := options.Find()
	opts.SetSort(bson.D{{Key: "_id", Value: -1}})
	if cursor, err = rn.client.Database("config").Collection("chunks").Find(ctx, filter, opts); err != nil {
		return err
	}
	i := 0
	for cursor.Next(ctx) {
		v := bson.M{}
		if err = cursor.Decode(&v); err != nil {
			continue
		}
		if v["shard"].(string) == otherShards[i]["_id"].(string) {
			i++
			continue
		}
		cmd := bson.D{{Key: "moveChunk", Value: ns}, {Key: "find", Value: v["min"].(bson.M)},
			{Key: "to", Value: otherShards[i]["_id"].(string)}}
		rn.Logger.Info(fmt.Sprintf("moving %v from %v to %v", v["min"], v["shard"], otherShards[i]["_id"]))
		if err = rn.client.Database("admin").RunCommand(ctx, cmd).Decode(&result); err != nil {
			log.Fatal(err)
		}
		i++
		if i == len(otherShards) {
			break
		}
	}
	time.Sleep(1 * time.Second)
	return nil
}
