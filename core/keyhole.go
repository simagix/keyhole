// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/globalsign/mgo"
)

// SimDBName - db name for simulation
var SimDBName = fmt.Sprintf("_KEYHOLE_%X", 1024+1024*rand.Intn(1024))

// CollectionName -
var CollectionName = "examples"

// Base -
type Base struct {
	dialInfo      *mgo.DialInfo
	uri           string
	ssl           bool
	sslCAFile     string
	sslPEMKeyFile string
	tps           int
	filename      string
	verbose       bool
	peek          bool
	monitor       bool
	bulkSize      int
	duration      int
	cleanup       bool
	drop          bool
	wmajor        bool
	dbName        string
}

var ssi MongoServerInfo

// NewBase - Constructor
func NewBase(dialInfo *mgo.DialInfo, uri string, ssl bool, sslCAFile string, sslPEMKeyFile string,
	tps int, filename string, verbose bool, peek bool, monitor bool,
	bulkSize int, duration int, cleanup bool, drop bool, wmajor bool, dbName string) Base {
	runner := Base{dialInfo, uri, ssl, sslCAFile, sslPEMKeyFile,
		tps, filename, verbose, peek, monitor,
		bulkSize, duration, cleanup, drop, wmajor, dbName}
	runner.initSimDocs()
	return runner
}

func getShardsURIList(session *mgo.Session, uri string) ([]string, error) {
	var uriList []string
	var err error
	if ssi, err = GetMongoServerInfo(session); err != nil {
		return uriList, err
	}
	uriList = append(uriList, uri)
	if ssi.Cluster == SHARDED {
		if uriList, err = GetShards(session, uri); err != nil {
			return uriList, err
		}
	}
	return uriList, err
}

// Start process requests
func (b Base) Start(session *mgo.Session, conn int, tx string, simonly bool) error {
	var err error
	var uriList []string
	if uriList, err = getShardsURIList(session, b.uri); err != nil {
		return err
	}
	log.Println("Duration in minute(s):", b.duration)
	b.terminationHandler(uriList, session)

	if b.peek == false && b.monitor == false { // keep --peek in case we need to hook to secondaries during load tests.
		if b.drop {
			Cleanup(session)
		}

		if err = ShardCollection(session); err != nil {
			return err
		}

		// Simulation mode
		// 1st minute - build up data and memory
		// 2nd and 3rd minutes - normal TPS ops
		// remaining minutes - burst with no delay
		// last minute - normal TPS ops until exit
		log.Printf("Total TPS: %d (tps) * %d (conns) = %d, duration: %d (mins), bulk size: %d\n",
			b.tps, conn, b.tps*conn, b.duration, b.bulkSize)

		tdoc := GetTransactions(tx)
		b.CreateIndexes(tdoc.Indexes)
		simTime := b.duration
		if simonly == false {
			simTime--
		}
		for i := 0; i < conn; i++ {
			go func() {
				if simonly == false {
					if err = b.PopulateData(b.wmajor); err != nil {
						return
					}
					time.Sleep(10 * time.Millisecond)
				}

				b.Simulate(simTime, tdoc.Transactions, b.wmajor)
			}()
		}
	}

	b.collectServerStatus(uriList, simonly)
	return err
}

func (b Base) terminationHandler(uriList []string, session *mgo.Session) {
	quit := make(chan os.Signal, 2)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	timer := time.NewTimer(time.Duration(b.duration) * time.Minute)

	go func(session *mgo.Session, uriList []string) {
		for {
			select {
			case <-quit:
				b.terminate(session, uriList)
			case <-timer.C:
				if b.monitor == false {
					b.terminate(session, uriList)
				}
			}
		}
	}(session, uriList)
}

func (b Base) terminate(session *mgo.Session, uriList []string) {
	var filenames []string
	var filename string
	var err error

	for _, value := range uriList {
		if filename, err = b.PrintServerStatus(value, 60); err != nil {
			continue
		}
		filenames = append(filenames, filename)
	}
	for _, filename := range filenames {
		log.Println("stats written to", filename)
	}
	if b.cleanup {
		Cleanup(session)
	}
	os.Exit(0)
}

func (b Base) collectServerStatus(uriList []string, simonly bool) {
	var channel = make(chan string)
	for _, value := range uriList {
		if b.monitor == false {
			if b.peek == true { // peek mode watch a defined db
				go b.CollectDBStats(value, channel, b.dbName)
			} else if simonly == false { // load test mode watches _KEYHOLE_88000
				go b.CollectDBStats(value, channel, SimDBName)
			}
		}
		go b.CollectServerStatus(value, channel)
	}

	// infinite loop waits for goroutine to send messages back
	for {
		msg := <-channel
		log.Print(msg)
		time.Sleep(time.Second * 1)
	}
}
