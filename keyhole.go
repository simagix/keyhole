// Copyright 2018 Kuei-chun Chen. All rights reserved.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/globalsign/mgo"
	keyhole "github.com/simagix/keyhole/core"
	"github.com/simagix/keyhole/mongo"
)

var version string

func main() {
	bulksize := flag.Int("bulksize", 100, "bulk insert size")
	cleanup := flag.Bool("cleanup", false, "clean up demo database")
	collection := flag.String("collection", "", "collection name to print schema")
	collscan := flag.Bool("collscan", false, "list only COLLSCAN (with --loginfo)")
	conn := flag.Int("conn", 10, "nuumber of connections")
	diag := flag.String("diag", "", "diagnosis of server status or diagnostic.data")
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
	span := flag.Int("span", -1, "granunarity for summary")
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

	var err error
	if *diag != "" {
		var str string
		var filenames = []string{*diag}
		if len(flag.Args()) > 0 {
			filenames = append(filenames, flag.Args()...)
		}
		tspan := *span
		if *webserver { // get data points summary if web server is enabled
			tspan = 300
		}
		d := keyhole.NewDiagnosticData(tspan)
		if str, err = d.PrintDiagnosticData(filenames, *webserver); err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
		fmt.Println(str)
		if *webserver {
			g := keyhole.NewGrafana(d)
			fmt.Println("Get more granular data points, data point every second.")
			go func(g *keyhole.Grafana) {
				d = keyhole.NewDiagnosticData(1)
				d.PrintDiagnosticData(filenames, *webserver)
				g.ReinitGrafana(d)
			}(g)
			keyhole.HTTPServer(5408, d, g)
		}
		os.Exit(0)
	} else if *loginfo != "" {
		if err = keyhole.LogInfo(*loginfo, *collscan); err != nil {
			fmt.Println(err)
		}
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
	} else if len(*uri) == 0 {
		fmt.Println("Missing connection string")
		fmt.Println("Usage: keyhole [opts] uri")
		flag.PrintDefaults()
		os.Exit(0)
	}

	var dialInfo *mgo.DialInfo

	if dialInfo, err = keyhole.ParseDialInfo(*uri); err != nil {
		panic(err)
	}

	if *verbose {
		log.Println("MongoDB URI:", *uri)
	}

	if dialInfo.Username != "" && dialInfo.Password == "" {
		if dialInfo.Password, err = keyhole.ReadPasswordFromStdin(); err != nil {
			panic(err)
		}
		index := strings.Index(*uri, "@")
		*uri = (*uri)[:index] + ":" + dialInfo.Password + (*uri)[index:]
	}

	session, err := keyhole.GetSessionWithTimeout(dialInfo, *wmajor, *ssl, *sslCAFile, *sslPEMKeyFile, 10*time.Second)
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
		fmt.Println(mongo.GetIndexes(session, dbName))
		os.Exit(0)
	}

	runner := keyhole.NewBase(dialInfo, *uri, *ssl, *sslCAFile, *sslPEMKeyFile,
		*tps, *file, *verbose, *peek, *monitor,
		*bulksize, *duration, *cleanup, *drop,
		*wmajor, dbName)
	if err = runner.Start(session, *conn, *tx, *simonly); err != nil {
		panic(err)
	}
}
