package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"

	"github.com/simagix/keyhole/stats"
	"github.com/simagix/keyhole/stats/mongo"
)

var x = 1024
var dbname = fmt.Sprintf("_KEYHOLE_%X", x+x*rand.Intn(x))

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
		bytes, _ := json.MarshalIndent(mongo.IsMaster(*uri), "", "  ")
		fmt.Println(string(bytes))
		os.Exit(0)
	} else if *seed == true {
		stats.Seed(*uri, *verbose)
		os.Exit(0)
	}

	// Simulation mode
	fmt.Printf("Total TPS: %d (tps) * %d (conns) = %d\n", *tps, *conn, *tps**conn)
	quit := make(chan os.Signal, 2)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-quit
		mongo.PrintServerStatus(*uri)
		mongo.Cleanup(*uri, dbname)
		os.Exit(0)
	}()

	mongo.Cleanup(*uri, dbname)
	go mongo.CollectServerStatus(*uri)
	go mongo.PrintDBStats(*uri, dbname)

	for i := 0; i < *conn; i++ {
		go func() {
			select {
			case <-quit:
				return
			default:
				sim := mongo.New(*uri, dbname, *tps)
				sim.PopulateData()
				sim.Simulate()
			}
		}()
	}

	var input string
	fmt.Println("Ctrl-C to quit...")
	fmt.Scanln(&input)
}
