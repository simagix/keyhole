// Copyright 2018 Kuei-chun Chen. All rights reserved.

package sim

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/simagix/keyhole/mdb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

// SimDBName - db name for simulation
var SimDBName = fmt.Sprintf("_KEYHOLE_%X", 1024+1024*rand.Intn(1024))

// CollectionName -
var CollectionName = "examples"

// Runner -
type Runner struct {
	uri           string
	sslCAFile     string
	sslPEMKeyFile string
	connString    connstring.ConnString
	client        *mongo.Client
	tps           int
	filename      string
	verbose       bool
	peek          bool
	monitor       bool
	duration      int
	cleanup       bool
	drop          bool
	conns         int
	txFilename    string
	simOnly       bool
}

var ssi mdb.ServerInfo

// NewRunner - Constructor
func NewRunner(uri string, sslCAFile string, sslPEMKeyFile string) (*Runner, error) {
	var err error
	var client *mongo.Client
	var runner Runner
	connString, _ := connstring.Parse(uri)

	if connString.Database == "" {
		connString.Database = mdb.KEYHOLEDB
		pos := strings.Index(uri, "?")
		if pos > 0 { // found ?query_string
			uri = (uri)[:pos] + connString.Database + (uri)[pos:]
		} else {
			length := len(uri)
			if (uri)[length-1] == '/' {
				uri += connString.Database
			} else {
				uri += "/" + connString.Database
			}
		}
	}

	if client, err = mdb.NewMongoClient(uri, sslCAFile, sslPEMKeyFile); err != nil {
		return &runner, err
	}
	runner = Runner{uri: uri, sslCAFile: sslCAFile, sslPEMKeyFile: sslPEMKeyFile,
		cleanup: true, connString: connString, client: client}
	runner.initSimDocs()
	return &runner, err
}

// SetTPS set transaction per second
func (rn *Runner) SetTPS(tps int) {
	rn.tps = tps
}

// SetTemplateFilename -
func (rn *Runner) SetTemplateFilename(filename string) {
	rn.filename = filename
}

// SetVerbose -
func (rn *Runner) SetVerbose(verbose bool) {
	rn.verbose = verbose
}

// SetPeekMode -
func (rn *Runner) SetPeekMode(mode bool) {
	rn.peek = mode
}

// SetMonitorMode -
func (rn *Runner) SetMonitorMode(mode bool) {
	rn.monitor = mode
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
	var uriList []string
	if uriList, err = mdb.GetShardsURIList(rn.client, rn.uri); err != nil {
		return err
	}
	ctx := context.Background()
	log.Println("Duration in minute(s):", rn.duration)
	rn.terminationHandler(uriList)

	if rn.peek == false && rn.monitor == false { // keep --peek in case we need to hook to secondaries during load tests.
		if rn.drop {
			rn.Cleanup()
		}

		var ssi mdb.ServerInfo
		if ssi, err = mdb.GetServerInfo(rn.client); err != nil {
			return err
		}

		if ssi.Cluster == mdb.SHARDED {
			collname := SimDBName + "." + CollectionName
			log.Println("Sharding collection:", collname)
			result := bson.M{}

			if err = rn.client.Database("admin").RunCommand(ctx, bson.D{{Key: "enableSharding", Value: SimDBName}}).Decode(&result); err != nil {
				return err
			}

			indexView := rn.client.Database(SimDBName).Collection(CollectionName).Indexes()
			idx := mongo.IndexModel{
				Keys: bson.D{{Key: "_id", Value: "hashed"}},
			}
			if _, err = indexView.CreateOne(ctx, idx); err != nil {
				return err
			}

			if err = rn.client.Database("admin").RunCommand(ctx, bson.D{{Key: "shardCollection", Value: collname}, {Key: "key", Value: bson.M{"_id": "hashed"}}}).Decode(&result); err != nil {
				return err
			}
		}

		// Simulation mode
		// 1st minute - build up data and memory
		// 2nd and 3rd minutes - normal TPS ops
		// remaining minutes - burst with no delay
		// last minute - normal TPS ops until exit
		log.Printf("Total TPS: %d (tps) * %d (conns) = %d, duration: %d (mins)\n",
			rn.tps, rn.conns, rn.tps*rn.conns, rn.duration)

		tdoc := GetTransactions(rn.txFilename)
		rn.CreateIndexes(tdoc.Indexes)
		simTime := rn.duration
		if rn.simOnly == false {
			simTime--
		}
		for i := 0; i < rn.conns; i++ {
			go func(thread int) {
				if rn.simOnly == false {
					if err = PopulateData(rn.uri, rn.sslCAFile, rn.sslPEMKeyFile); err != nil {
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
	}

	rn.collectAllStatus(uriList)
	return err
}

func (rn *Runner) terminationHandler(uriList []string) {
	quit := make(chan os.Signal, 2)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	timer := time.NewTimer(time.Duration(rn.duration) * time.Minute)

	go func(uriList []string) {
		for {
			select {
			case <-quit:
				rn.terminate(uriList)
			case <-timer.C:
				if rn.monitor == false {
					rn.terminate(uriList)
				}
			}
		}
	}(uriList)
}

func (rn *Runner) terminate(uriList []string) {
	var filenames []string
	var filename string
	var err error

	if rn.cleanup {
		rn.Cleanup()
	}
	for _, uri := range uriList {
		if filename, err = rn.PrintServerStatus(uri, 60); err != nil {
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

func (rn *Runner) collectAllStatus(uriList []string) {
	var channel = make(chan string)
	var err error
	var client *mongo.Client

	for _, uri := range uriList {
		if client, err = mdb.NewMongoClient(uri, rn.sslCAFile, rn.sslPEMKeyFile); err != nil {
			log.Println(err)
			continue
		}
		if rn.monitor == false {
			if rn.peek == true { // peek mode watch a defined db
				go rn.CollectDBStats(client, channel, rn.connString.Database, uri)
			} else if rn.simOnly == false { // load test mode watches _KEYHOLE_88000
				go rn.CollectDBStats(client, channel, SimDBName, uri)
			}
		}
		go rn.ReplSetGetStatus(uri, channel)
		go rn.CollectServerStatus(uri, channel)
	}

	// infinite loop waits for goroutine to send messages back
	for {
		msg := <-channel
		log.Print(msg)
		time.Sleep(time.Second * 1)
	}
}

// CreateIndexes creates indexes
func (rn *Runner) CreateIndexes(docs []bson.M) error {
	var err error
	var ctx = context.Background()
	c := rn.client.Database(SimDBName).Collection(CollectionName)
	indexView := c.Indexes()

	if len(docs) == 0 {
		idx := mongo.IndexModel{
			Keys: bson.D{{Key: "favoriteCity", Value: 1}},
		}
		if _, err = indexView.CreateOne(ctx, idx); err != nil {
			return err
		}
	}
	idx := mongo.IndexModel{
		Keys: bson.D{{Key: "_search", Value: 1}},
	}
	if _, err = indexView.CreateOne(ctx, idx); err != nil {
		return err
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
	ctx := context.Background()
	log.Println("dropping collection", SimDBName, CollectionName)
	if err = rn.client.Database(SimDBName).Collection(CollectionName).Drop(ctx); err != nil {
		log.Println(err)
	}
	log.Println("dropping database", SimDBName)
	if err = rn.client.Database(SimDBName).Drop(ctx); err != nil {
		log.Println(err)
	}
	return err
}
