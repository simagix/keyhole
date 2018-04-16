package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/simagix/keyhole/stats"
)

func main() {
	uri := flag.String("uri", "mongodb://localhost", "MongoDB URI")
	info := flag.Bool("info", false, "get cluster info")
	seed := flag.Bool("seed", false, "seed a database for demo")
	conn := flag.Int("conn", 20, "nuumber of connections")
	tps := flag.Int("tps", 600, "number of trasaction per second per connection")
	verbose := flag.Bool("v", false, "verbose")

	flag.Parse()
	fmt.Println("MongoDB URI:", *uri)

	if *info == true {
		bytes, _ := json.MarshalIndent(stats.IsMaster(*uri), "", "  ")
		fmt.Println(string(bytes))
		os.Exit(0)
	} else if *seed == true {
		stats.Seed(*uri, *verbose)
		os.Exit(0)
	}

	// Simulation mode
	fmt.Printf("Total TPS: %d (tps) * %d (conns) = %d\n", *tps, *conn, *tps**conn)
	m := stats.New(*uri, stats.DBName, *tps)
	m.Cleanup()
	go m.CollectServerStatus()
	go m.PrintDBStats()

	quit := make(chan os.Signal, 2)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-quit
		m.PrintServerStatus()
		m.Cleanup()
		os.Exit(0)
	}()

	for i := 0; i < *conn; i++ {
		go func() {
			select {
			case <-quit:
				return
			default:
				m := stats.New(*uri, stats.DBName, *tps)
				m.PopulateData()
				m.Simulate()
			}
		}()
	}

	var input string
	fmt.Println("Ctrl-C to quit...")
	fmt.Scanln(&input)
}
