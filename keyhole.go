// Copyright 2018 Kuei-chun Chen. All rights reserved.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/simagix/keyhole/stats"
	mgo "gopkg.in/mgo.v2"
)

var version string

func main() {
	bulksize := flag.Int("bulksize", 10, "bulk insert size")
	conn := flag.Int("conn", 20, "nuumber of connections")
	duration := flag.Int("duration", 5, "load test duration in minutes")
	drop := flag.Bool("drop", false, "drop examples collection before seeding")
	file := flag.String("file", "", "template file for seedibg data")
	info := flag.Bool("info", false, "get cluster info")
	loginfo := flag.String("loginfo", "", "log performance analytic")
	monitor := flag.Bool("monitor", false, "collects server status every 10 minutes for 24 hours")
	nocleanup := flag.Bool("nocleanup", false, "keep keyhole demo database")
	peek := flag.Bool("peek", false, "only collect stats")
	quote := flag.Bool("quote", false, "print a quote")
	quotes := flag.Bool("quotes", false, "print all quotes")
	schema := flag.Bool("schema", false, "print schema")
	seed := flag.Bool("seed", false, "seed a database for demo")
	simonly := flag.Bool("simonly", false, "simulation only mode")
	ssl := flag.Bool("ssl", false, "use TLS/SSL")
	sslCA := flag.String("sslCAFile", "", "CA file")
	tps := flag.Int("tps", 600, "number of trasaction per second per connection")
	total := flag.Int("total", 1000, "nuumber of documents to create")
	tx := flag.String("tx", "", "file with defined transactions")
	uri := flag.String("uri", "", "MongoDB URI")
	ver := flag.Bool("version", false, "print version number")
	verbose := flag.Bool("v", false, "verbose")
	view := flag.String("view", "", "server status file")
	wmajor := flag.Bool("wmajor", false, "{w: majority}")

	flag.Parse()

	flagset := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { flagset[f.Name] = true })
	// if --monitor is set, it collects server status every 10 minutes.
	// Unless --duration is set, it changes duration to 1440 minutes (24 hours)
	if *monitor {
		if flagset["duration"] == false {
			*duration = 1440
		}
	}

	if *quote {
		stats.PrintQuote()
		os.Exit(0)
	} else if *quotes {
		stats.PrintQuotes()
		os.Exit(0)
	} else if *view != "" {
		stats.AnalyzeServerStatus(*view)
		os.Exit(0)
	} else if *loginfo != "" {
		stats.LogInfo(*loginfo)
		os.Exit(0)
	} else if *ver {
		fmt.Println("keyhole ver.", version)
		os.Exit(0)
	} else if *schema {
		if *file == "" {
			bytes, _ := json.MarshalIndent(stats.GetRandomDoc(), "", "  ")
			doc := strings.Replace(string(bytes), "mongodb.", "", -1)
			doc = strings.Replace(doc, "simagix.", "", -1)
			fmt.Println(doc)
		} else {
			bytes, _ := json.MarshalIndent(stats.GetDocByTemplate(*file, false), "", "  ")
			fmt.Println(string(bytes))
		}
		os.Exit(0)
	} else if len(*uri) == 0 {
		fmt.Println("Missing --uri")
		flag.PrintDefaults()
		os.Exit(0)
	}
	fmt.Println("MongoDB URI:", *uri)
	dialInfo, _ := mgo.ParseURL(*uri)
	dbName := dialInfo.Database
	if dialInfo.Database == "" {
		dbName = "_KEYHOLE_"
	}

	if *info == true {
		session, err := stats.GetSession(*uri, *ssl, *sslCA)
		if err != nil {
			panic(err)
		}
		bytes, _ := json.MarshalIndent(stats.ServerInfo(session), "", "  ")
		fmt.Println(string(bytes))
		session.Close()
		os.Exit(0)
	} else if *seed == true {
		session, err := stats.GetSession(*uri, *ssl, *sslCA)
		if err != nil {
			panic(err)
		}
		defer session.Close()
		if *file == "" {
			stats.Seed(session, *drop, dbName, *verbose)
		} else {
			stats.SeedFromTemplate(session, *file, *total, *drop, dbName, *verbose)
		}
		os.Exit(0)
	}

	session, err := stats.GetSession(*uri, *ssl, *sslCA)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	ssi := stats.ServerInfo(session)
	var uriList []string
	uriList = append(uriList, *uri)
	if ssi.Cluster == "sharded" {
		list := stats.GetShards(session, *uri)
		uriList = list
	}

	fmt.Println("Duration in minute(s):", *duration)
	m := stats.New(*uri, *ssl, *sslCA, *tps, *file, *verbose, !*nocleanup, *peek, *monitor, *bulksize)
	timer := time.NewTimer(time.Duration(*duration) * time.Minute)
	quit := make(chan os.Signal, 2)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-quit:
				for _, value := range uriList {
					m.PrintServerStatus(value)
				}
				os.Exit(0)
			case <-timer.C:
				for _, value := range uriList {
					m.PrintServerStatus(value)
				}
				m.Cleanup()
				os.Exit(0)
			}
		}
	}()

	if *peek == false && *monitor == false { // keep --peek in case we need to hook to secondaries during load tests.
		m.Cleanup()

		if ssi.Cluster == "sharded" {
			stats.ShardCollection(session)
		}

		if *wmajor {
			fmt.Println("{w : majority}")
		}
		if *tps < *bulksize {
			*bulksize = *tps
		}
		// Simulation mode
		// 1st minute - build up data and memory
		// 2nd and 3rd minutes - normal TPS ops
		// remaining minutes - burst with no delay
		// last minute - normal TPS ops until exit
		fmt.Printf("Total TPS: %d (tps) * %d (conns) = %d, duration: %d (mins), bulk size: %d\n",
			*tps, *conn, *tps**conn, *duration, *bulksize)
		m.CreateIndexes()
		simTime := *duration
		if *simonly == false {
			simTime--
		}
		for i := 0; i < *conn; i++ {
			go func() {
				if *simonly == false {
					m.PopulateData(*wmajor)
					time.Sleep(1 * time.Second)
				}
				m.Simulate(simTime, *tx, *wmajor)
			}()
		}
	}

	var channel = make(chan string)
	for _, value := range uriList {
		if *monitor == false {
			if *peek == true { // peek mode watch a defined db
				go m.CollectDBStats(value, channel, dbName)
			} else { // load test mode watches _KEYHOLE_88000
				go m.CollectDBStats(value, channel, stats.SimDBName)
			}
		}
		go m.CollectServerStatus(value, channel)
	}

	// infinite loop waits for goroutine to send messages back
	for {
		msg := <-channel
		fmt.Print(msg)
		time.Sleep(time.Second * 1)
	}
}
