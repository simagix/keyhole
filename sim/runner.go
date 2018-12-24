// Copyright 2018 Kuei-chun Chen. All rights reserved.

package sim

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/x/bsonx"
	"github.com/mongodb/mongo-go-driver/x/network/connstring"
	"github.com/simagix/keyhole/mdb"
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
	tps           int
	filename      string
	verbose       bool
	peek          bool
	monitor       bool
	duration      int
	cleanup       bool
	drop          bool
	connString    connstring.ConnString
	client        *mongo.Client
}

var ssi mdb.ServerInfo

// NewRunner - Constructor
func NewRunner(uri string, sslCAFile string, sslPEMKeyFile string, tps int, filename string,
	verbose bool, peek bool, monitor bool, duration int, cleanup bool, drop bool) Runner {
	var err error
	var client *mongo.Client
	connString, _ := connstring.Parse(uri)
	if client, err = mdb.NewMongoClient(uri, sslCAFile, sslPEMKeyFile); err != nil {
		panic(err)
	}
	runner := Runner{uri, sslCAFile, sslPEMKeyFile, tps, filename,
		verbose, peek, monitor, duration, cleanup, drop, connString, client}
	runner.initSimDocs()
	return runner
}

// Start process requests
func (rn Runner) Start(client *mongo.Client, conn int, tx string, simonly bool) error {
	var err error
	var uriList []string
	if uriList, err = mdb.GetShardsURIList(client, rn.uri); err != nil {
		return err
	}
	ctx := context.Background()
	log.Println("Duration in minute(s):", rn.duration)
	rn.terminationHandler(uriList, client)

	if rn.peek == false && rn.monitor == false { // keep --peek in case we need to hook to secondaries during load tests.
		if rn.drop {
			Cleanup(client)
		}

		var ssi mdb.ServerInfo
		if ssi, err = mdb.GetServerInfo(client); err != nil {
			return err
		}

		if ssi.Cluster == mdb.SHARDED {
			collname := SimDBName + "." + CollectionName
			log.Println("Sharding collection:", collname)
			result := bson.M{}

			if err = client.Database("admin").RunCommand(ctx, bson.D{{Key: "enableSharding", Value: SimDBName}}).Decode(&result); err != nil {
				return err
			}

			indexView := client.Database(SimDBName).Collection(CollectionName).Indexes()
			idx := mongo.IndexModel{
				Keys: bsonx.Doc{{Key: "_id", Value: bsonx.String("hashed")}},
			}
			if _, err = indexView.CreateOne(ctx, idx); err != nil {
				return err
			}

			if err = client.Database("admin").RunCommand(ctx, bson.D{{Key: "shardCollection", Value: collname}, {Key: "key", Value: bson.M{"_id": "hashed"}}}).Decode(&result); err != nil {
				return err
			}
		}

		// Simulation mode
		// 1st minute - build up data and memory
		// 2nd and 3rd minutes - normal TPS ops
		// remaining minutes - burst with no delay
		// last minute - normal TPS ops until exit
		log.Printf("Total TPS: %d (tps) * %d (conns) = %d, duration: %d (mins)\n",
			rn.tps, conn, rn.tps*conn, rn.duration)

		tdoc := GetTransactions(tx)
		rn.CreateIndexes(tdoc.Indexes)
		simTime := rn.duration
		if simonly == false {
			simTime--
		}
		for i := 0; i < conn; i++ {
			go func() {
				if simonly == false {
					if err = PopulateData(rn.uri, rn.sslCAFile, rn.sslPEMKeyFile); err != nil {
						panic(err)
					}
					time.Sleep(10 * time.Millisecond)
				}

				rn.Simulate(simTime, tdoc.Transactions)
			}()
		}
	}

	rn.collectAllStatus(uriList, simonly)
	return err
}

func (rn Runner) terminationHandler(uriList []string, client *mongo.Client) {
	quit := make(chan os.Signal, 2)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	timer := time.NewTimer(time.Duration(rn.duration) * time.Minute)

	go func(client *mongo.Client, uriList []string) {
		for {
			select {
			case <-quit:
				rn.terminate(client, uriList)
			case <-timer.C:
				if rn.monitor == false {
					rn.terminate(client, uriList)
				}
			}
		}
	}(client, uriList)
}

func (rn Runner) terminate(client *mongo.Client, uriList []string) {
	var filenames []string
	var filename string
	var err error

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
	if rn.cleanup {
		Cleanup(client)
	}
	os.Exit(0)
}

func (rn Runner) collectAllStatus(uriList []string, simonly bool) {
	var channel = make(chan string)
	var err error
	var client *mongo.Client

	for _, uri := range uriList {
		if client, err = mdb.NewMongoClient(uri, rn.sslCAFile, rn.sslPEMKeyFile); err != nil {
			continue
		}
		if err = client.Connect(context.Background()); err != nil {
			continue
		}
		// TODO if err = mdb.AddCertificates(dialInfo, rn.sslCAFile, rn.sslPEMKeyFile); err != nil {
		// 	continue
		// }
		if rn.monitor == false {
			if rn.peek == true { // peek mode watch a defined db
				go rn.CollectDBStats(client, channel, rn.connString.Database, uri)
			} else if simonly == false { // load test mode watches _KEYHOLE_88000
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
func (rn Runner) CreateIndexes(docs []bson.M) error {
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err = rn.client.Connect(ctx); err != nil {
		panic(err)
	}
	defer rn.client.Disconnect(ctx)
	c := rn.client.Database(SimDBName).Collection(CollectionName)
	indexView := c.Indexes()

	if len(docs) == 0 {
		idx := mongo.IndexModel{
			Keys: bsonx.Doc{{Key: "favoriteCity", Value: bsonx.Int32(1)}},
		}
		indexView.CreateOne(ctx, idx)
	}
	idx := mongo.IndexModel{
		Keys: bsonx.Doc{{Key: "_search", Value: bsonx.Int32(1)}},
	}
	indexView.CreateOne(ctx, idx)

	for _, doc := range docs {
		keys := bsonx.Doc{}
		fmt.Println(doc)
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

			keys = append(keys, bsonx.Elem{Key: k, Value: bsonx.Int32(int32(x))})
		}
		idx := mongo.IndexModel{
			Keys: keys,
		}
		indexView.CreateOne(ctx, idx)
	}

	return err
}

// Cleanup drops the temp database
func Cleanup(client *mongo.Client) error {
	var err error
	ctx := context.Background()
	log.Println("dropping collection", SimDBName, CollectionName)
	client.Database(SimDBName).Collection(CollectionName).Drop(ctx)
	log.Println("dropping database", SimDBName)
	client.Database(SimDBName).Drop(ctx)
	return err
}
