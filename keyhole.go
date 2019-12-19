// Copyright 2018 Kuei-chun Chen. All rights reserved.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/simagix/gox"
	"github.com/simagix/keyhole/mdb"
	"github.com/simagix/keyhole/sim"
	"github.com/simagix/keyhole/sim/util"
	"github.com/simagix/mongo-atlas/atlas"
	anly "github.com/simagix/mongo-ftdc/analytics"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

var version = "self-built"

func main() {
	caFile := flag.String("sslCAFile", "", "CA file")
	changeStreams := flag.Bool("changeStreams", false, "change streams watch")
	clientPEMFile := flag.String("sslPEMKeyFile", "", "client PEM file")
	collection := flag.String("collection", "", "collection name to print schema")
	collscan := flag.Bool("collscan", false, "list only COLLSCAN (with --loginfo)")
	cardinality := flag.String("cardinality", "", "check collection cardinality")
	conn := flag.Int("conn", 10, "nuumber of connections")
	diag := flag.String("diag", "", "diagnosis of server status or diagnostic.data")
	doodle := flag.Bool("doodle", false, "print random values of sample docs")
	duration := flag.Int("duration", 5, "load test duration in minutes")
	drop := flag.Bool("drop", false, "drop examples collection before seeding")
	explain := flag.String("explain", "", "explain a query from a JSON doc or a log line")
	file := flag.String("file", "", "template file for seedibg data")
	ftdc := flag.String("ftdc", "", "process FTDC and output processed data")
	index := flag.Bool("index", false, "get indexes info")
	info := flag.Bool("info", false, "get cluster info | Atlas info (atlas://user:key)")
	loginfo := flag.String("loginfo", "", "log performance analytic")
	peek := flag.Bool("peek", false, "only collect stats")
	pipe := flag.String("pipeline", "", "aggregation pipeline")
	regex := flag.String("regex", "", "regex pattern for loginfo")
	schema := flag.Bool("schema", false, "print schema")
	seed := flag.Bool("seed", false, "seed a database for demo")
	simonly := flag.Bool("simonly", false, "simulation only mode")
	span := flag.Int("span", -1, "granunarity for summary")
	tps := flag.Int("tps", 300, "number of trasaction per second per connection")
	total := flag.Int("total", 1000, "nuumber of documents to create")
	tx := flag.String("tx", "", "file with defined transactions")
	uri := flag.String("uri", "", "MongoDB URI") // orverides connection uri from args
	ver := flag.Bool("version", false, "print version number")
	verbose := flag.Bool("v", false, "verbose")
	webserver := flag.Bool("web", false, "enable web server")

	flag.Parse()
	if *uri == "" && len(flag.Args()) > 0 {
		*uri = flag.Arg(0)
	}
	flagset := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { flagset[f.Name] = true })
	var err error
	if *ftdc != "" {
		filenames := append([]string{*ftdc}, flag.Args()...)
		if *webserver == true {
			anly.SingleJSONServer(filenames)
		} else {
			met := anly.NewMetrics(filenames)
			met.SetOutputOnly(true)
			met.Read()
			os.Exit(0)
		}
	} else if *diag != "" {
		filenames := append([]string{*diag}, flag.Args()...)
		if *webserver == true { // backward compatible
			anly.SingleJSONServer(filenames)
		} else {
			metrics := anly.NewDiagnosticData(*span)
			if str, e := metrics.PrintDiagnosticData(filenames); e != nil {
				log.Fatal(e)
			} else {
				fmt.Println(str)
			}
		}
		os.Exit(0)
	} else if *info == true && strings.Index(*uri, "atlas://") == 0 {
		var api *atlas.API
		if api, err = atlas.ParseURI(*uri); err != nil {
			log.Fatal(err)
		}
		api.SetVerbose(*verbose)
		var str string
		if str, err = api.GetClustersSummary(); err != nil {
			log.Fatal(err)
		}
		fmt.Println(str)
		os.Exit(0)
	} else if strings.Index(*loginfo, "atlas://") == 0 {
		var atl *atlas.Log
		if atl, err = atlas.ParseLogURI(*loginfo); err != nil {
			log.Fatal(err)
		}
		atl.SetVerbose(*verbose)
		var filenames []string
		if filenames, err = atl.Download(); err != nil {
			log.Fatal(err)
		}
		for _, filename := range filenames {
			fmt.Println("=> processing", filename)
			var str string
			li := mdb.NewLogInfo(filename)
			li.SetVerbose(*verbose)
			if str, err = li.Analyze(); err != nil {
				log.Println(err)
				continue
			}
			fmt.Println(str)
		}
		os.Exit(0)
	} else if *loginfo != "" {
		var str string
		li := mdb.NewLogInfo(*loginfo)
		li.SetRegexPattern(*regex)
		li.SetCollscan(*collscan)
		li.SetVerbose(*verbose)
		if str, err = li.Analyze(); err != nil {
			log.Fatal(err)
		}
		fmt.Println(str)
		if li.OutputFilename != "" {
			log.Println("Encoded output written to", li.OutputFilename)
		}
		os.Exit(0)
	} else if *ver {
		fmt.Println("keyhole", version)
		os.Exit(0)
	} else if *schema && *uri == "" {
		if *file == "" {
			fmt.Println(util.GetDemoSchema())
		} else {
			fmt.Println(util.GetDemoFromFile(*file))
		}
		os.Exit(0)
	} else if *explain != "" && *uri == "" { //--explain file.json.gz (w/o uri)
		exp := mdb.NewExplain()
		if err = exp.PrintExplainResults(*explain); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	} else if len(*uri) == 0 {
		fmt.Println("Missing connection string")
		fmt.Println("Usage: keyhole [opts] uri")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *uri, err = mdb.Parse(*uri); err != nil {
		log.Fatal(err)
	}

	client, err := mdb.NewMongoClient(*uri, *caFile, *clientPEMFile)
	if err != nil {
		log.Fatal(err)
	}
	connString, err := connstring.Parse(*uri)
	if err != nil {
		log.Fatal(err)
	}

	if *info == true {
		mc := mdb.NewMongoCluster(client)
		mc.SetVerbose(*verbose)
		mc.SetHost(connString.Hosts[0])
		mc.SetDoodleMode(*doodle)
		if doc, e := mc.GetClusterInfo(); e != nil {
			log.Fatal(e)
		} else if *verbose == false {
			fmt.Println(gox.Stringify(doc, "", "  "))
		}
		os.Exit(0)
	} else if *seed == true {
		f := sim.NewFeeder()
		f.SetCollection(*collection)
		f.SetDatabase(connString.Database)
		f.SetFile(*file)
		f.SetIsDrop(*drop)
		f.SetTotal(*total)
		if err = f.SeedData(client); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	} else if *index == true {
		ir := mdb.NewIndexesReader(client)
		if connString.Database == mdb.KEYHOLEDB {
			connString.Database = ""
		}
		ir.SetDBName(connString.Database)
		ir.SetVerbose(*verbose)
		m, e := ir.GetIndexes()
		if e != nil {
			log.Fatal(e)
		}
		ir.Print(m)
		os.Exit(0)
	} else if *schema == true {
		var str string
		if str, err = sim.GetSchemaFromCollection(client, connString.Database, *collection); err != nil {
			log.Fatal(err)
		}
		fmt.Println(str)
		os.Exit(0)
	} else if *cardinality != "" { // --card <collection> [-v]
		card := mdb.NewCardinality(client)
		card.SetVerbose(*verbose)
		if summary, e := card.GetCardinalityArray(connString.Database, *cardinality); e != nil {
			log.Fatal(e)
		} else {
			fmt.Println(card.GetSummary(summary))
		}
		os.Exit(0)
	} else if *explain != "" { // --explain json_or_log_file  [-v]
		exp := mdb.NewExplain()
		exp.SetVerbose(*verbose)
		if err = exp.ExecuteAllPlans(client, *explain); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	} else if *changeStreams == true {
		stream := mdb.NewChangeStream()
		stream.SetCollection(*collection)
		stream.SetDatabase(connString.Database)
		stream.SetPipelineString(*pipe)
		stream.Watch(client, util.Echo)
		os.Exit(0)
	}

	client.Disconnect(context.Background())
	var runner *sim.Runner
	if runner, err = sim.NewRunner(*uri, *caFile, *clientPEMFile); err != nil {
		log.Fatal(err)
	}
	runner.SetTPS(*tps)
	runner.SetTemplateFilename(*file)
	runner.SetVerbose(*verbose)
	runner.SetPeekingMode(*peek)
	runner.SetSimulationDuration(*duration)
	runner.SetDropFirstMode(*drop)
	runner.SetNumberConnections(*conn)
	runner.SetTransactionTemplateFilename(*tx)
	runner.SetSimOnlyMode(*simonly)
	if err = runner.Start(); err != nil {
		log.Fatal(err)
	}
	runner.CollectAllStatus()
}
