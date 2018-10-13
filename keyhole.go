// Copyright 2018 Kuei-chun Chen. All rights reserved.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/simagix/keyhole/charts"
	keyhole "github.com/simagix/keyhole/core"
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

	if *diag != "" {
		keyhole.AnalyzeServerStatus(*diag, *span, false)
		os.Exit(0)
	} else if *loginfo != "" {
		keyhole.LogInfo(*loginfo, *collscan, *verbose)
		os.Exit(0)
	} else if *ver {
		fmt.Println("keyhole ver.", version)
		os.Exit(0)
	} else if *schema && *uri == "" {
		if *file == "" {
			fmt.Println(keyhole.GetDemoSchema())
		} else {
			fmt.Println(keyhole.GetDemoFromFile(*file))
		}
		os.Exit(0)
	} else if *webserver && *file != "" {
		keyhole.AnalyzeServerStatus(*file, 10, true)
		charts.HTTPServer(5408)
	} else if len(*uri) == 0 {
		fmt.Println("Missing connection string")
		fmt.Println("Usage: keyhole [opts] uri")
		flag.PrintDefaults()
		os.Exit(0)
	}

	dialInfo, err := keyhole.ParseDialInfo(*uri)
	if err != nil {
		panic(err)
	}

	if *verbose {
		fmt.Println("MongoDB URI:", *uri)
	}

	session, err := keyhole.GetSession(dialInfo, *ssl, *sslCAFile, *sslPEMKeyFile)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	dbName := dialInfo.Database
	if dialInfo.Database == "" && *index == false {
		dbName = "_KEYHOLE_"
	}

	if *info == true {
		var info keyhole.MongoServerInfo
		if info, err = keyhole.GetMongoServerInfo(session); err != nil {
			panic(err)
		}
		bytes, _ := json.MarshalIndent(info, "", "  ")
		fmt.Println(string(bytes))
		os.Exit(0)
	} else if *seed == true {
		sb := keyhole.NewSeedBase(*file, *collection, *total, *drop, dbName, *verbose)
		if err = sb.SeedData(session); err != nil {
			fmt.Println(err)
		}
		os.Exit(0)
	} else if *schema == true {
		var str string
		if str, err = keyhole.GetSchemaFromCollection(session, dbName, *collection, *verbose); err != nil {
			fmt.Println(err)
		}
		fmt.Println(str)
		os.Exit(0)
	} else if *index == true {
		fmt.Println(keyhole.GetIndexes(session, dbName, *verbose))
		os.Exit(0)
	}

	runner := keyhole.NewBase(dialInfo, *uri, *ssl, *sslCAFile, *sslPEMKeyFile,
		*tps, *file, *verbose, *peek, *monitor,
		*bulksize, *duration, *span, *cleanup, *drop,
		*wmajor, dbName)
	runner.Start(session, *conn, *tx, *simonly)
}
