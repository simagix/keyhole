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
	conn := flag.Int("conn", 20, "nuumber of connections")
	nocleanup := flag.Bool("nocleanup", false, "keep keyhole demo database")
	duration := flag.Int("duration", 5, "load test duration in minutes")
	drop := flag.Bool("drop", false, "drop examples collection before seeding")
	file := flag.String("file", "", "template file for seedibg data")
	info := flag.Bool("info", false, "get cluster info")
	loginfo := flag.String("loginfo", "", "log performance analytic")
	peek := flag.Bool("peek", false, "only collect data")
	quote := flag.Bool("quote", false, "print a quote")
	quotes := flag.Bool("quotes", false, "print all quotes")
	schema := flag.Bool("schema", false, "print schema")
	seed := flag.Bool("seed", false, "seed a database for demo")
	simonly := flag.Bool("simonly", false, "simulation only mode")
	ssl := flag.Bool("ssl", false, "use TLS/SSL")
	sslCA := flag.String("sslCAFile", "", "CA file")
	tps := flag.Int("tps", 600, "number of trasaction per second per connection")
	uri := flag.String("uri", "", "MongoDB URI")
	ver := flag.Bool("version", false, "print version number")
	verbose := flag.Bool("v", false, "verbose")
	view := flag.String("view", "", "server status file")

	flag.Parse()
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
			stats.SeedFromTemplate(session, *file, *drop, *verbose)
		}
		os.Exit(0)
	}

	session, err := stats.GetSession(*uri, *ssl, *sslCA)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	ssi := stats.ServerInfo(session)

	curi := *uri
	if ssi.Cluster == "sharded" {
		list := stats.GetShards(session, *uri)
		curi = list[0]
	}

	m := stats.New(*uri, *ssl, *sslCA, stats.DBName, *tps, *file, *verbose, !*nocleanup)
	timer := time.NewTimer(time.Duration(*duration) * time.Minute)
	quit := make(chan os.Signal, 2)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-quit
		m.PrintServerStatus(curi)
		m.Cleanup()
		os.Exit(0)
	}()
	go func() {
		<-timer.C
		m.PrintServerStatus(curi)
		m.Cleanup()
		os.Exit(0)
	}()

	if *peek == false {
		m.Cleanup()

		if ssi.Cluster == "sharded" {
			stats.ShardCollection(session, stats.DBName+"."+stats.CollectionName)
		}

		// Simulation mode
		// 1st minute - build up data and memory
		// 2nd and 3rd minutes - normal TPS ops
		// remaining minutes - burst with no delay
		// last minute - normal TPS ops until exit
		go m.PrintDBStats()
		fmt.Printf("Total TPS: %d (tps) * %d (conns) = %d, duration = %d (mins)\n", *tps, *conn, *tps**conn, *duration)
		m.CreateIndexes()
		for i := 0; i < *conn; i++ {
			go func() {
				select {
				case <-quit:
					return
				default:
					msim := stats.New(*uri, *ssl, *sslCA, stats.DBName, *tps, *file, *verbose, !*nocleanup)
					if *simonly == false {
						msim.PopulateData()
						*duration--
					}
					msim.Simulate(*duration)
					time.Sleep(time.Second)
				}
			}()
		}
	}

	m.CollectServerStatus(curi)
}
