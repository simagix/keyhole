// Copyright 2018 Kuei-chun Chen. All rights reserved.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/simagix/keyhole/mdb"
	"github.com/simagix/keyhole/mdb/atlas"
	"github.com/simagix/keyhole/sim"
	"github.com/simagix/keyhole/sim/util"
	"github.com/simagix/keyhole/web"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/x/network/connstring"
)

var version = "self-built"

func main() {
	caFile := flag.String("sslCAFile", "", "CA file")
	changeStreams := flag.Bool("changeStreams", false, "change streams watch")
	clientPEMFile := flag.String("sslPEMKeyFile", "", "client PEM file")
	collection := flag.String("collection", "", "collection name to print schema")
	collscan := flag.Bool("collscan", false, "list only COLLSCAN (with --loginfo)")
	card := flag.Bool("card", false, "check collection cardinality")
	conn := flag.Int("conn", 10, "nuumber of connections")
	diag := flag.String("diag", "", "diagnosis of server status or diagnostic.data")
	duration := flag.Int("duration", 5, "load test duration in minutes")
	drop := flag.Bool("drop", false, "drop examples collection before seeding")
	file := flag.String("file", "", "template file for seedibg data")
	index := flag.Bool("index", false, "get indexes info")
	info := flag.Bool("info", false, "get cluster info | Atlas info (atlas://user:key)")
	loginfo := flag.String("loginfo", "", "log performance analytic")
	monitor := flag.Bool("monitor", false, "collects server status every 10 seconds")
	peek := flag.Bool("peek", false, "only collect stats")
	pipe := flag.String("pipeline", "", "aggregation pipeline")
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
		d := sim.NewDiagnosticData(tspan)
		if str, err = d.PrintDiagnosticData(filenames, *webserver); err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
		fmt.Println(str)
		if *webserver {
			g := web.NewGrafana()
			g.SetFTDCSummaryStats(d)

			fmt.Println("Get more granular data points, data point every second.")
			go func(g *web.Grafana, d *sim.DiagnosticData, filenames []string) {
				d = sim.NewDiagnosticData(1)
				d.PrintDiagnosticData(filenames, true)
				g.SetFTDCDetailStats(d)
			}(g, d, filenames)
			web.HTTPServer(5408, d, g)
		}
		os.Exit(0)
	} else if *info == true && strings.Index(*uri, "atlas://") == 0 {
		var str string
		idx := strings.Index(*uri, "@")
		if idx > 0 {
			*uri = (*uri)[:idx]
		}
		*uri = (*uri)[8:]
		su := atlas.NewSummary(*uri)
		su.SetVerbose(*verbose)
		if str, err = su.GetSummary(); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(str)
		}
		os.Exit(0)
	} else if strings.Index(*loginfo, "atlas://") == 0 {
		lg := atlas.ParseAtlasURI(*loginfo)
		lg.SetVerbose(*verbose)
		if lg.Error() != "" {
			panic(lg.Error())
		}
		var filenames []string
		if filenames, err = lg.DownloadLogs("."); err != nil {
			panic(err)
		}
		for _, filename := range filenames {
			fmt.Println("=> processing", filename)
			var str string
			li := util.NewLogInfo(filename)
			li.SetVerbose(*verbose)
			if str, err = li.Analyze(); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(str)
			}
		}
		os.Exit(0)
	} else if *loginfo != "" {
		var str string
		li := util.NewLogInfo(*loginfo)
		li.SetCollscan(*collscan)
		li.SetVerbose(*verbose)
		if str, err = li.Analyze(); err != nil {
			log.Fatal(err)
		}
		fmt.Println(str)
		log.Println("Encoded output written to", li.OutputFilename)
		os.Exit(0)
	} else if *ver {
		fmt.Println("keyhole ver.", version)
		os.Exit(0)
	} else if *schema && *uri == "" {
		if *file == "" {
			fmt.Println(util.GetDemoSchema())
		} else {
			fmt.Println(util.GetDemoFromFile(*file))
		}
		os.Exit(0)
	} else if len(*uri) == 0 {
		fmt.Println("Missing connection string")
		fmt.Println("Usage: keyhole [opts] uri")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *uri, err = mdb.Parse(*uri); err != nil {
		panic(err)
	}

	client, err := mdb.NewMongoClient(*uri, *caFile, *clientPEMFile)
	if err != nil {
		panic(err)
	}
	connString, err := connstring.Parse(*uri)
	if err != nil {
		panic(err)
	}

	if *info == true {
		mc := mdb.NewMongoCluster(client)
		mc.SetVerbose(*verbose)
		doc, e := mc.GetClusterInfo()
		if e != nil {
			panic(e)
		}
		if *verbose == true {
			if err = mc.WriteJSON(connString.Hosts[0] + ".json"); err != nil {
				panic(err)
			}
			fmt.Println("JSON is written to", connString.Hosts[0]+".json")
		} else {
			fmt.Println(mdb.Stringify(doc, "", "  "))
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
			fmt.Println(err)
		}
		os.Exit(0)
	} else if *index == true {
		if connString.Database == mdb.KEYHOLEDB {
			connString.Database = ""
		}
		fmt.Println(mdb.GetIndexes(client, connString.Database))
		os.Exit(0)
	} else if *schema == true {
		var str string
		if str, err = sim.GetSchemaFromCollection(client, connString.Database, *collection); err != nil {
			fmt.Println(err)
		}
		fmt.Println(str)
		os.Exit(0)
	} else if *card == true {
		card := mdb.NewCardinality(connString.Database, *collection)
		card.SetVerbose(*verbose)
		doc, e := card.CheckCardinality(client)
		if e != nil {
			fmt.Println(e)
		} else {
			fmt.Println(mdb.Stringify(doc, "", "   "))
		}
		os.Exit(0)
	} else if *changeStreams == true {
		var pipeline = []bson.D{}
		if *pipe != "" {
			pipeline = mdb.MongoPipeline(*pipe)
		}
		stream := mdb.NewChangeStream()
		stream.SetCollection(*collection)
		stream.SetDatabase(connString.Database)
		stream.SetPipeline(pipeline)
		stream.Watch(client, util.Echo)
		os.Exit(0)
	}

	client.Disconnect(context.Background())
	var runner *sim.Runner
	if runner, err = sim.NewRunner(*uri, *caFile, *clientPEMFile); err != nil {
		panic(err)
	}
	runner.SetTPS(*tps)
	runner.SetTemplateFilename(*file)
	runner.SetVerbose(*verbose)
	runner.SetPeekMode(*peek)
	runner.SetMonitorMode(*monitor)
	runner.SetSimulationDuration(*duration)
	runner.SetDropFirstMode(*drop)
	runner.SetNumberConnections(*conn)
	runner.SetTransactionTemplateFilename(*tx)
	runner.SetSimOnlyMode(*simonly)
	if err = runner.Start(); err != nil {
		panic(err)
	}
}
