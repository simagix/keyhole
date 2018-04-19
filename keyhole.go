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
	uri := flag.String("uri", "mongodb://localhost", "MongoDB URI")
	info := flag.Bool("info", false, "get cluster info")
	seed := flag.Bool("seed", false, "seed a database for demo")
	conn := flag.Int("conn", 20, "nuumber of connections")
	tps := flag.Int("tps", 600, "number of trasaction per second per connection")
	duration := flag.Int("duration", 6, "load test duration in minutes")
	verbose := flag.Bool("v", false, "verbose")
	peek := flag.Bool("peek", false, "only collect data")
	view := flag.String("view", "", "server status file")

	flag.Parse()
	fmt.Println("MongoDB URI:", *uri)

	if *info == true {
		bytes, _ := json.MarshalIndent(stats.IsMaster(*uri), "", "  ")
		fmt.Println(string(bytes))
		os.Exit(0)
	} else if *seed == true {
		stats.Seed(*uri, *verbose)
		os.Exit(0)
	} else if *view != "" {
		stats.AnalyzeServerStatus(*view)
		os.Exit(0)
	}

	// Simulation mode
	// 1st minute - build up data and memory
	// 2nd and 3rd minutes - normal TPS ops
	// remaining minutes - burst with no delay
	// last minute - normal TPS ops until exit
	m := stats.New(*uri, stats.DBName, *tps)
	go m.PrintDBStats()
	timer := time.NewTimer(time.Duration(*duration) * time.Minute)
	quit := make(chan os.Signal, 2)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-quit
		cleanup(m)
	}()
	go func() {
		<-timer.C
		cleanup(m)
	}()

	if *peek == false {
		fmt.Printf("Total TPS: %d (tps) * %d (conns) = %d, duration = %d (mins)\n", *tps, *conn, *tps**conn, *duration)
		m.Cleanup()
		for i := 0; i < *conn; i++ {
			go func() {
				select {
				case <-quit:
					return
				default:
					msim := stats.New(*uri, stats.DBName, *tps)
					msim.PopulateData()
					msim.Simulate(*duration)
				}
			}()
		}
	}

	m.CollectServerStatus()
}

func cleanup(m stats.MongoConn) {
	m.PrintServerStatus()
	m.Cleanup()
	os.Exit(0)
}
