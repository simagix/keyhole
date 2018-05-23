package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/simagix/keyhole/stats"
)

func main() {
	uri := flag.String("uri", "", "MongoDB URI")
	info := flag.Bool("info", false, "get cluster info")
	seed := flag.Bool("seed", false, "seed a database for demo")
	conn := flag.Int("conn", 20, "nuumber of connections")
	tps := flag.Int("tps", 600, "number of trasaction per second per connection")
	duration := flag.Int("duration", 6, "load test duration in minutes")
	verbose := flag.Bool("v", false, "verbose")
	peek := flag.Bool("peek", false, "only collect data")
	view := flag.String("view", "", "server status file")
	ssl := flag.Bool("ssl", false, "use TLS/SSL")
	sslCA := flag.String("sslCAFile", "", "CA file")

	flag.Parse()
	if len(*uri) == 0 {
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
		stats.Seed(session, *verbose)
		session.Close()
		os.Exit(0)
	} else if *view != "" {
		stats.AnalyzeServerStatus(*view)
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

	m := stats.New(*uri, *ssl, *sslCA, stats.DBName, *tps)
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
		for i := 0; i < *conn; i++ {
			go func() {
				select {
				case <-quit:
					return
				default:
					msim := stats.New(*uri, *ssl, *sslCA, stats.DBName, *tps)
					msim.PopulateData()
					msim.Simulate(*duration)
				}
			}()
		}
	}

	m.CollectServerStatus(curi)
}
