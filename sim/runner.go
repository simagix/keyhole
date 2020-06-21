// Copyright 2018 Kuei-chun Chen. All rights reserved.

package sim

import (
	"bufio"
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/simagix/gox"
	"github.com/simagix/keyhole/mdb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

// Runner -
type Runner struct {
	auto                  bool
	channel               chan string
	client                *mongo.Client
	collectionName        string
	conns                 int
	dbName                string
	drop                  bool
	duration              int
	filename              string
	metrics               map[string][]bson.M
	mutex                 sync.RWMutex
	peek                  bool
	serverInfo            mdb.ServerInfo
	simOnly               bool
	tlsCAFile             string
	tlsCertificateKeyFile string
	tps                   int
	txFilename            string
	uri                   string
	uriList               []string
	verbose               bool
}

// NewRunner - Constructor
func NewRunner(uri string, tlsCAFile string, tlsCertificateKeyFile string) (*Runner, error) {
	var err error
	runner := Runner{tlsCAFile: tlsCAFile, tlsCertificateKeyFile: tlsCertificateKeyFile,
		channel: make(chan string), collectionName: "examples", metrics: map[string][]bson.M{},
		mutex: sync.RWMutex{}}
	connString, _ := connstring.Parse(uri)
	runner.dbName = connString.Database
	if connString.Database == "" {
		runner.dbName = mdb.KEYHOLEDB
		pos := strings.Index(uri, "?")
		if pos > 0 { // found ?query_string
			uri = (uri)[:pos] + runner.dbName + (uri)[pos:]
		} else {
			length := len(uri)
			if (uri)[length-1] == '/' {
				uri += runner.dbName
			} else {
				uri += "/" + runner.dbName
			}
		}
	}
	if runner.client, err = mdb.NewMongoClient(uri, tlsCAFile, tlsCertificateKeyFile); err != nil {
		return &runner, err
	}
	if runner.serverInfo, err = mdb.GetServerInfo(runner.client); err != nil {
		return &runner, err
	}
	runner.uriList = []string{uri}
	if runner.serverInfo.Cluster == mdb.SHARDED {
		if runner.uriList, err = mdb.GetShardListWithURI(runner.client, uri); err != nil {
			return &runner, err
		}
	}
	runner.uri = runner.uriList[len(runner.uriList)-1]
	return &runner, err
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

// SetVerbose -
func (rn *Runner) SetVerbose(verbose bool) {
	rn.verbose = verbose
}

// SetPeekingMode -
func (rn *Runner) SetPeekingMode(mode bool) {
	rn.peek = mode
	if rn.peek == true {
		go func(x int) {
			time.Sleep(time.Duration(x) * time.Minute)
			rn.terminate()
		}(rn.duration)
	}
}

// SetSimulationDuration -
func (rn *Runner) SetSimulationDuration(duration int) {
	rn.duration = duration
}

// SetDropFirstMode -
func (rn *Runner) SetDropFirstMode(mode bool) {
	rn.drop = mode
}

// SetNumberConnections -
func (rn *Runner) SetNumberConnections(num int) {
	rn.conns = num
}

// SetTransactionTemplateFilename -
func (rn *Runner) SetTransactionTemplateFilename(filename string) {
	rn.txFilename = filename
}

// SetSimOnlyMode -
func (rn *Runner) SetSimOnlyMode(mode bool) {
	rn.simOnly = mode
}

// Start process requests
func (rn *Runner) Start() error {
	var err error
	if rn.peek == true {
		return nil
	}
	if rn.auto == false {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Begin a load test [Y/n]: ")
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		if text != "y" && text != "" && text != "Y" {
			os.Exit(0)
		}
	}
	log.Println("Duration in minute(s):", rn.duration)
	if rn.drop {
		rn.Cleanup()
	}
	rn.initSimDocs()
	tdoc := GetTransactions(rn.txFilename)
	rn.createIndexes(tdoc.Indexes)

	// Simulation mode
	// 1st minute - build up data and memory
	// 2nd and 3rd minutes - normal TPS ops
	// remaining minutes - burst with no delay
	// last minute - normal TPS ops until exit
	log.Printf("Total TPS: %d (%d tps/conn * %d conns), duration: %d (mins)\n", rn.tps*rn.conns, rn.tps, rn.conns, rn.duration)
	simTime := rn.duration
	if rn.simOnly == false {
		simTime--
	}
	for i := 0; i < rn.conns; i++ {
		go func(thread int) {
			if rn.simOnly == false && rn.duration > 0 {
				if err = rn.PopulateData(); err != nil {
					log.Println("Thread", thread, "existing with", err)
					return
				}
				time.Sleep(10 * time.Millisecond)
			}

			if err = rn.Simulate(simTime, tdoc.Transactions, thread); err != nil {
				log.Println("Thread", thread, "existing with", err)
				return
			}
		}(i)
	}
	return nil
}

func (rn *Runner) terminate() {
	var client *mongo.Client
	var filenames []string
	var filename string
	var err error

	rn.Cleanup()
	for _, uri := range rn.uriList {
		if client, err = mdb.NewMongoClient(uri, rn.tlsCAFile, rn.tlsCertificateKeyFile); err != nil {
			log.Println(err)
			continue
		}
		stats := NewServerStats(uri, rn.channel)
		stats.SetVerbose(rn.verbose)
		if filename, err = stats.printServerStatus(client); err != nil {
			log.Println(err)
			continue
		}
		filenames = append(filenames, filename)
	}
	for _, filename := range filenames {
		log.Println("stats written to", filename)
	}
	os.Exit(0)
}

// CollectAllStatus collects all server stats
func (rn *Runner) CollectAllStatus() error {
	var err error
	for i, uri := range rn.uriList {
		var client *mongo.Client
		if client, err = mdb.NewMongoClient(uri, rn.tlsCAFile, rn.tlsCertificateKeyFile); err != nil {
			log.Println(err)
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
			log.Print(<-rn.channel)
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

		if rn.serverInfo.Cluster == mdb.SHARDED {
			if err = rn.splitChunks(); err != nil {
				fmt.Println(err)
			}
		}
	}

	for _, doc := range docs {
		keys := bson.D{}
		for k, v := range doc {
			x := int32(1)
			switch v.(type) {
			case int:
				if v.(int) < 0 {
					x = -1
				}
			case float64:
				if v.(float64) < 0 {
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
	if rn.peek == false {
		ctx := context.Background()
		log.Println("dropping collection", rn.dbName, rn.collectionName)
		if err = rn.client.Database(rn.dbName).Collection(rn.collectionName).Drop(ctx); err != nil {
			log.Println(err)
		}
		log.Println("dropping database", rn.dbName)
		if err = rn.client.Database(rn.dbName).Drop(ctx); err != nil {
			log.Println(err)
		}
		filename := "keyhole_perf." + fileTimestamp + ".enc.gz"
		var data bytes.Buffer
		gob.Register(time.Duration(0))
		enc := gob.NewEncoder(&data)
		if err = enc.Encode(rn.metrics); err != nil {
			log.Println("encode error:", err)
		}
		gox.OutputGzipped(data.Bytes(), filename)
		log.Println("optime written to", filename)
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
	log.Println("Sharding collection:", ns)
	cmd := bson.D{{Key: "enableSharding", Value: rn.dbName}}
	if err = rn.client.Database("admin").RunCommand(ctx, cmd).Decode(&result); err != nil {
		return err
	}
	cmd = bson.D{{Key: "shardCollection", Value: ns}, {Key: "key", Value: bson.M{"email": 1}}}
	if err = rn.client.Database("admin").RunCommand(ctx, cmd).Decode(&result); err != nil {
		return err
	}
	log.Println("splitting chunks...")
	if cursor, err = rn.client.Database("config").Collection("shards").Find(ctx, bson.D{{}}); err != nil {
		return err
	}
	shards := []bson.M{}
	for cursor.Next(ctx) {
		v := bson.M{}
		if err = cursor.Decode(&v); err != nil {
			log.Println(err)
			continue
		}
		if primary != v["_id"].(string) {
			shards = append(shards, v)
		}
	}
	shardKeys := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M",
		"N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
	divider := 1 + len(shardKeys)/(len(shards)+1)
	for i := range shards {
		cmd := bson.D{{Key: "split", Value: ns}, {Key: "middle", Value: bson.M{"email": shardKeys[(i+1)*divider]}}}
		if err = rn.client.Database("admin").RunCommand(ctx, cmd).Decode(&result); err != nil {
			log.Println(err)
		}
	}

	log.Println("moving chunks...")
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
		if v["shard"].(string) == shards[i]["_id"].(string) {
			i++
			continue
		}
		cmd := bson.D{{Key: "moveChunk", Value: ns}, {Key: "find", Value: v["min"].(bson.M)},
			{Key: "to", Value: shards[i]["_id"].(string)}}
		log.Printf("moving %v from %v to %v\n", v["min"], v["shard"], shards[i]["_id"])
		if err = rn.client.Database("admin").RunCommand(ctx, cmd).Decode(&result); err != nil {
			log.Fatal(err)
		}
		i++
		if i == len(shards) {
			break
		}
	}
	time.Sleep(1 * time.Second)
	return nil
}
