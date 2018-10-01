// Copyright 2018 Kuei-chun Chen. All rights reserved.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/simagix/keyhole/charts"
	"github.com/simagix/keyhole/stats"
	"github.com/simagix/keyhole/utils"
	"golang.org/x/crypto/ssh/terminal"
)

var version string

func main() {
	bulksize := flag.Int("bulksize", 512, "bulk insert size")
	cleanup := flag.Bool("cleanup", false, "clean up demo database")
	collection := flag.String("collection", "", "collection name to print schema")
	collscan := flag.Bool("collscan", false, "list only COLLSCAN (with --loginfo)")
	conn := flag.Int("conn", 10, "nuumber of connections")
	diag := flag.String("diag", "", "diagnosis of server status")
	duration := flag.Int("duration", 5, "load test duration in minutes")
	drop := flag.Bool("drop", false, "drop examples collection before seeding")
	file := flag.String("file", "", "template file for seedibg data")
	index := flag.Bool("index", false, "get indexes info")
	info := flag.Bool("info", false, "get cluster info")
	loginfo := flag.String("loginfo", "", "log performance analytic")
	monitor := flag.Bool("monitor", false, "collects server status every 10 seconds")
	peek := flag.Bool("peek", false, "only collect stats")
	quote := flag.Bool("quote", false, "print a quote")
	quotes := flag.Bool("quotes", false, "print all quotes")
	schema := flag.Bool("schema", false, "print schema")
	seed := flag.Bool("seed", false, "seed a database for demo")
	simonly := flag.Bool("simonly", false, "simulation only mode")
	span := flag.Int("span", 60, "granunarity for summary")
	ssl := flag.Bool("ssl", false, "use TLS/SSL")
	sslCAFile := flag.String("sslCAFile", "", "CA file")
	sslPEMKeyFile := flag.String("sslPEMKeyFile", "", "client PEM file")
	tps := flag.Int("tps", 300, "number of trasaction per second per connection")
	total := flag.Int("total", 1000, "nuumber of documents to create")
	tx := flag.String("tx", "", "file with defined transactions")
	uri := flag.String("uri", "", "MongoDB URI") // orverides connection uri from args
	ver := flag.Bool("version", false, "print version number")
	verbose := flag.Bool("v", false, "verbose")
	webserver := flag.Bool("web", false, "enable web server")
	wmajor := flag.Bool("wmajor", false, "{w: majority}")

	flag.Parse()
	if *uri == "" && len(flag.Args()) > 0 {
		*uri = flag.Arg(0)
	}
	flagset := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { flagset[f.Name] = true })

	if *ssl && (len(*sslCAFile) == 0 || len(*sslPEMKeyFile) == 0) {
		if len(*sslCAFile) != len(*sslPEMKeyFile) {
			panic("need both CA and PEM files")
		}
	}

	if *quote {
		utils.PrintQuote()
		os.Exit(0)
	} else if *quotes {
		utils.PrintQuotes()
		os.Exit(0)
	} else if *diag != "" {
		stats.AnalyzeServerStatus(*diag, *span, false)
		os.Exit(0)
	} else if *loginfo != "" {
		stats.LogInfo(*loginfo, *collscan)
		os.Exit(0)
	} else if *ver {
		fmt.Println("keyhole ver.", version)
		os.Exit(0)
	} else if *schema && *uri == "" {
		if *file == "" {
			fmt.Println(stats.GetDemoSchema())
		} else {
			fmt.Println(stats.GetDemoFromFile(*file))
		}
		os.Exit(0)
	} else if *webserver && *file != "" {
		stats.AnalyzeServerStatus(*file, 10, true)
		charts.HTTPServer(5408)
	} else if len(*uri) == 0 {
		fmt.Println("Missing connection string")
		fmt.Println("Usage: keyhole [opts] uri")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *verbose {
		fmt.Println("MongoDB URI:", *uri)
	}

	dialInfo, err := stats.ParseDialInfo(*uri)
	if err != nil {
		panic(err)
	}

	if dialInfo.Username != "" && dialInfo.Password == "" && (runtime.GOOS == "darwin" || runtime.GOOS == "linux") {
		fmt.Print("Enter Password: ")
		bytePassword, _ := terminal.ReadPassword(int(syscall.Stdin))
		dialInfo.Password = string(bytePassword)
		fmt.Println("")
	}

	dbName := dialInfo.Database
	if dialInfo.Database == "" && *index == false {
		dbName = "_KEYHOLE_"
	}

	session, err := stats.GetSession(dialInfo, *ssl, *sslCAFile, *sslPEMKeyFile)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	if *info == true {
		bytes, _ := json.MarshalIndent(stats.GetMongoServerInfo(session), "", "  ")
		fmt.Println(string(bytes))
		os.Exit(0)
	} else if *seed == true {
		if *file == "" {
			stats.Seed(session, *total, *drop, dbName, *verbose)
		} else {
			if *collection == "" {
				fmt.Println("usage: keyhole --uri connection_uri --seed [--file filename --collection collection_name]")
				os.Exit(0)
			}
			stats.SeedFromTemplate(session, *file, *collection, *total, *drop, dbName, *verbose)
		}
		os.Exit(0)
	} else if *schema == true {
		if *collection == "" {
			fmt.Println("usage: keyhole --schema [--file filename | --uri connection_uri --collection collection_name]")
			os.Exit(0)
		}
		fmt.Println(stats.GetSchemaFromCollection(session, dbName, *collection, *verbose))
		os.Exit(0)
	} else if *index == true {
		fmt.Println(stats.GetIndexes(session, dbName, *verbose))
		os.Exit(0)
	}

	ssi := stats.GetMongoServerInfo(session)
	var uriList []string
	uriList = append(uriList, *uri)
	if ssi.Cluster == "sharded" {
		uriList = stats.GetShards(session, *uri)
	}

	if *webserver || *monitor == true { // web server enabled will be in monitor mode
		*monitor = true
		*duration = 0
		if *webserver {
			go charts.HTTPServer(5408)
		}
	} else {
		fmt.Println("Duration in minute(s):", *duration)
	}
	m := stats.New(dialInfo, *uri, *ssl, *sslCAFile, *sslPEMKeyFile, *tps, *file, *verbose, *peek, *monitor, *bulksize)
	timer := time.NewTimer(time.Duration(*duration) * time.Minute)
	quit := make(chan os.Signal, 2)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-quit:
				for _, value := range uriList {
					m.PrintServerStatus(value, *span)
				}
				if *cleanup {
					m.Cleanup()
				}
				os.Exit(0)
			case <-timer.C:
				if *monitor == false {
					for _, value := range uriList {
						m.PrintServerStatus(value, *span)
					}
					if *cleanup {
						m.Cleanup()
					}
					os.Exit(0)
				}
			}
		}
	}()

	if *peek == false && *monitor == false { // keep --peek in case we need to hook to secondaries during load tests.
		if *drop {
			m.Cleanup()
		}

		if ssi.Cluster == "sharded" {
			stats.ShardCollection(session)
		}

		if *wmajor {
			fmt.Println("{w : majority}")
		}
		// Simulation mode
		// 1st minute - build up data and memory
		// 2nd and 3rd minutes - normal TPS ops
		// remaining minutes - burst with no delay
		// last minute - normal TPS ops until exit
		fmt.Printf("Total TPS: %d (tps) * %d (conns) = %d, duration: %d (mins), bulk size: %d\n",
			*tps, *conn, *tps**conn, *duration, *bulksize)

		tdoc := stats.GetTransactions(*tx)
		m.CreateIndexes(tdoc.Indexes)
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

				m.Simulate(simTime, tdoc.Transactions, *wmajor)
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
