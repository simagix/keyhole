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
	verbose := flag.Bool("v", false, "verbose")

	flag.Parse()
	fmt.Println("MongoDB URI:", *uri)
	fmt.Println("info:", *info)
	fmt.Println("seed:", *seed)

	if *info == true {
		bytes, _ := json.MarshalIndent(mongo.IsMaster(*uri), "", "  ")
		fmt.Println(string(bytes))
		os.Exit(0)
	} else if *seed == true {
		stats.Seed(*uri, *verbose)
		os.Exit(0)
	}

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		if *verbose {
			mongo.PrintServerStatus()
		}
		mongo.Cleanup(*uri, dbname)
		os.Exit(0)
	}()

	mongo.Cleanup(*uri, dbname)
	go mongo.CollectServerStatus(*uri)
	go mongo.CollectDBStats(*uri, dbname)

	for i := 0; i < 10; i++ {
		go mongo.Insert(*uri, dbname)
	}
	var input string
	fmt.Println("Ctrl-C to quit...")
	fmt.Scanln(&input)
}
