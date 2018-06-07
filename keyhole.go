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
)

var version string

func main() {
	bulksize := flag.Int("bulksize", 10, "bulk insert size")
	conn := flag.Int("conn", 20, "nuumber of connections")
	nocleanup := flag.Bool("nocleanup", false, "keep keyhole demo database")
	duration := flag.Int("duration", 5, "load test duration in minutes")
	drop := flag.Bool("drop", false, "drop examples collection before seeding")
	file := flag.String("file", "", "template file for seedibg data")
	info := flag.Bool("info", false, "get cluster info")
	loginfo := flag.String("loginfo", "", "log performance analytic")
	monitor := flag.Bool("monitor", false, "collects server status every 10 minutes for 24 hours")
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
			stats.Seed(session, *verbose)
		} else {
			stats.SeedFromTemplate(session, *file, *total, *drop, *verbose)
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
	m := stats.New(*uri, *ssl, *sslCA, stats.DBName, *tps, *file, *verbose, !*nocleanup, *peek, *monitor, *bulksize)
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
			stats.ShardCollection(session, stats.DBName+"."+stats.CollectionName)
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
		fmt.Printf("Total TPS: %d (tps) * %d (conns) = %d, duration: %d (mins), bulk size: %d\n", *tps, *conn, *tps**conn, *duration, *bulksize)
		m.CreateIndexes()
		simTime := *duration
		if *simonly == false {
			simTime--
		}
		for i := 0; i < *conn; i++ {
			go func() {
				// msim := stats.New(*uri, *ssl, *sslCA, stats.DBName, *tps, *file, *verbose, !*nocleanup, *peek, *monitor)
				if *simonly == false {
					m.PopulateData(*wmajor)
					time.Sleep(time.Second)
				}
				m.Simulate(simTime, *wmajor)
				time.Sleep(time.Second)
			}()
		}
	}

	var channel = make(chan string)
	for _, value := range uriList {
		if *monitor == false {
			go m.CollectDBStats(value, channel)
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
