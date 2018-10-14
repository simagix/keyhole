// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"fmt"
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
	span          int
	cleanup       bool
	drop          bool
	wmajor        bool
	dbName        string
}

var ssi MongoServerInfo

// NewBase - Constructor
func NewBase(dialInfo *mgo.DialInfo, uri string, ssl bool, sslCAFile string, sslPEMKeyFile string,
	tps int, filename string, verbose bool, peek bool, monitor bool,
	bulkSize int, duration int, span int, cleanup bool, drop bool,
	wmajor bool, dbName string) Base {
	runner := Base{dialInfo, uri, ssl, sslCAFile, sslPEMKeyFile,
		tps, filename, verbose, peek, monitor,
		bulkSize, duration, span, cleanup, drop,
		wmajor, dbName}
	runner.initSimDocs()
	return runner
}

func getShardsURIList(session *mgo.Session, uri string) ([]string, error) {
	var uriList []string
	var err error
	ssi, err = GetMongoServerInfo(session)
	if err != nil {
		return uriList, err
	}
	uriList = append(uriList, uri)
	if ssi.Cluster == SHARDED {
		var e error
		uriList, e = GetShards(session, uri)
		if e != nil {
			return uriList, err
		}
	}
	return uriList, err
}

// Start process requests
func (b Base) Start(session *mgo.Session, conn int, tx string, simonly bool) {
	var err error
	uriList, err := getShardsURIList(session, b.uri)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Duration in minute(s):", b.duration)
	b.printStats(uriList)

	if b.peek == false && b.monitor == false { // keep --peek in case we need to hook to secondaries during load tests.
		if b.drop {
			b.Cleanup()
		}

		if err = ShardCollection(session); err != nil {
			fmt.Println(err)
			return
		}

		if b.wmajor {
			fmt.Println("{w : majority}")
		}
		// Simulation mode
		// 1st minute - build up data and memory
		// 2nd and 3rd minutes - normal TPS ops
		// remaining minutes - burst with no delay
		// last minute - normal TPS ops until exit
		fmt.Printf("Total TPS: %d (tps) * %d (conns) = %d, duration: %d (mins), bulk size: %d\n",
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
						panic(err)
					}
					time.Sleep(10 * time.Millisecond)
				}

				b.Simulate(simTime, tdoc.Transactions, b.wmajor)
			}()
		}
	}

	b.collectServerStatus(uriList, simonly)
}

func (b Base) printStats(uriList []string) {
	quit := make(chan os.Signal, 2)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	timer := time.NewTimer(time.Duration(b.duration) * time.Minute)

	go func() {
		for {
			select {
			case <-quit:
				for _, value := range uriList {
					b.PrintServerStatus(value, b.span)
				}
				if b.cleanup {
					b.Cleanup()
				}
				os.Exit(0)
			case <-timer.C:
				if b.monitor == false {
					for _, value := range uriList {
						b.PrintServerStatus(value, b.span)
					}
					if b.cleanup {
						b.Cleanup()
					}
					os.Exit(0)
				}
			}
		}
	}()
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
		fmt.Print(msg)
		time.Sleep(time.Second * 1)
	}
}
