// Copyright 2018 Kuei-chun Chen. All rights reserved.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
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
	changeStreams := flag.Bool("changeStreams", false, "change streams watch")
	collection := flag.String("collection", "", "collection name to print schema")
	collscan := flag.Bool("collscan", false, "list only COLLSCAN (with --loginfo)")
	cardinality := flag.String("cardinality", "", "check collection cardinality")
	conn := flag.Int("conn", runtime.NumCPU(), "nuumber of connections")
	diag := flag.String("diag", "", "diagnosis of server status or diagnostic.data")
	doodle := flag.Bool("doodle", false, "print random values of sample docs")
	duration := flag.Int("duration", 5, "load test duration in minutes")
	drop := flag.Bool("drop", false, "drop examples collection before seeding")
	explain := flag.String("explain", "", "explain a query from a JSON doc or a log line")
	file := flag.String("file", "", "template file for seedibg data")
	ftdc := flag.Bool("ftdc", false, "download from atlas://user:key@group/cluster")
	index := flag.Bool("index", false, "get indexes info")
	info := flag.Bool("info", false, "get cluster info | Atlas info (atlas://user:key)")
	loginfo := flag.Bool("loginfo", false, "log performance analytic from file or Atlas")
	peek := flag.Bool("peek", false, "only collect stats")
	pause := flag.Bool("pause", false, "pause an Atlas cluster atlas://user:key@group/cluster")
	pipe := flag.String("pipeline", "", "aggregation pipeline")
	port := flag.Int("port", 5408, "web server port number")
	regex := flag.String("regex", "", "regex pattern for loginfo")
	request := flag.String("request", "", "Atlas API command")
	resume := flag.Bool("resume", false, "resume an Atlas cluster atlas://user:key@group/cluster")
	schema := flag.Bool("schema", false, "print schema")
	seed := flag.Bool("seed", false, "seed a database for demo")
	simonly := flag.Bool("simonly", false, "simulation only mode")
	span := flag.Int("span", -1, "granunarity for summary")
	sslCAFile := flag.String("sslCAFile", "", "CA file")
	sslPEMKeyFile := flag.String("sslPEMKeyFile", "", "client PEM file")
	tlsCAFile := flag.String("tlsCAFile", "", "TLS CA file")
	tlsCertificateKeyFile := flag.String("tlsCertificateKeyFile", "", "TLS CertificateKey File")
	tps := flag.Int("tps", 20, "number of trasaction per second per connection")
	total := flag.Int("total", 1000, "nuumber of documents to create")
	tx := flag.String("tx", "", "file with defined transactions")
	uri := flag.String("uri", "", "MongoDB URI") // orverides connection uri from args
	ver := flag.Bool("version", false, "print version number")
	verbose := flag.Bool("v", false, "verbose")
	vv := flag.Bool("vv", false, "very verbose")
	webserver := flag.Bool("web", false, "enable web server")
	wt := flag.Bool("wt", false, "visualize wiredTiger cache usage")
	yes := flag.Bool("yes", false, "bypass confirmation")

	flag.Parse()
	if *tlsCAFile == "" && *sslCAFile != "" {
		*tlsCAFile = *sslCAFile
	}
	if *tlsCertificateKeyFile == "" && *sslPEMKeyFile != "" {
		*tlsCertificateKeyFile = *sslPEMKeyFile
	}
	if *uri == "" && len(flag.Args()) > 0 {
		*uri = flag.Arg(0)
	}
	flagset := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { flagset[f.Name] = true })
	var err error
	if *conn < 1 {
		*conn = runtime.NumCPU()
	}
	if strings.HasPrefix(*uri, "atlas://") {
		var api *atlas.API
		if api, err = atlas.ParseURI(*uri); err != nil {
			log.Fatal(err)
		}
		api.SetArgs(flag.Args())
		api.SetFTDC(*ftdc)
		api.SetInfo(*info)
		api.SetLoginfo(*loginfo)
		api.SetPause(*pause)
		api.SetResume(*resume)
		api.SetRequest(*request)
		api.SetVerbose(*verbose)
		fmt.Println(api.Execute())

		if *loginfo {
			for _, filename := range api.GetLogNames() {
				fmt.Println("=> processing", filename)
				var str string
				li := mdb.NewLogInfo()
				li.SetVerbose(*verbose)
				if str, err = li.Analyze(filename); err != nil {
					log.Println(err)
					continue
				}
				fmt.Println(str)
			}
		}
		os.Exit(0)
	} else if *diag != "" {
		filenames := append([]string{*diag}, flag.Args()...)
		if *webserver == true { // backward compatible
			metrics := anly.NewMetrics()
			metrics.ProcessFiles(filenames)
			addr := fmt.Sprintf(":%d", *port)
			log.Fatal(http.ListenAndServe(addr, nil))
		} else {
			metrics := anly.NewDiagnosticData(*span)
			if str, e := metrics.PrintDiagnosticData(filenames); e != nil {
				log.Fatal(e)
			} else {
				fmt.Println(str)
			}
		}
		os.Exit(0)
	} else if *webserver {
		metrics := anly.NewMetrics()
		metrics.ProcessFiles(flag.Args())
		addr := fmt.Sprintf(":%d", *port)
		log.Fatal(http.ListenAndServe(addr, nil))
	} else if *loginfo {
		if len(flag.Args()) < 1 {
			log.Fatal("Usage: keyhole --loginfo filename")
		}
		filenames := []string{}
		for i, arg := range flag.Args() { // backward compatible
			if arg == "-collscan" || arg == "--collscan" {
				*collscan = true
			} else if arg == "-v" || arg == "--v" {
				*verbose = true
			} else if (arg == "-regex" || arg == "--regex") && *regex != "" {
				*regex = flag.Args()[i+1]
			} else {
				filenames = append(filenames, arg)
			}
		}
		li := mdb.NewLogInfo()
		li.SetRegexPattern(*regex)
		li.SetCollscan(*collscan)
		li.SetVerbose(*verbose)
		for _, filename := range filenames {
			var str string
			if str, err = li.Analyze(filename); err != nil {
				log.Fatal(err)
			}
			fmt.Println(str)
			if li.OutputFilename != "" {
				log.Println("Encoded output written to", li.OutputFilename)
			}
		}
		os.Exit(0)
	} else if *ver {
		fmt.Println("keyhole", version)
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

	client, err := mdb.NewMongoClient(*uri, *tlsCAFile, *tlsCertificateKeyFile)
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
		mc.SetVeryVerbose(*vv)
		mc.SetConnString(connString)
		mc.SetDoodleMode(*doodle)
		if doc, e := mc.GetClusterInfo(); e != nil {
			log.Fatal(e)
		} else if *verbose == false && *vv == false {
			fmt.Println(gox.Stringify(doc, "", "  "))
		}
		os.Exit(0)
	} else if *seed == true {
		f := sim.NewFeeder()
		f.SetCollection(*collection)
		f.SetConnections(*conn)
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
		if *collection == "" {
			log.Fatal("usage: keyhole [-v] --schema --collection collection_name <mongodb_uri>")
		}
		c := client.Database(connString.Database).Collection(*collection)
		var str string
		if str, err = sim.GetSchema(c, *verbose); err != nil {
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

	go func() {
		http.HandleFunc("/", gox.Cors(handler))
		addr := fmt.Sprintf(":%d", *port)
		log.Println(http.ListenAndServe(addr, nil))
	}()
	if *wt == true {
		wtc := mdb.NewWiredTigerCache(client)
		log.Printf("URL: http://localhost:%d/wt\n", *port)
		wtc.Start()
	}

	var runner *sim.Runner
	if runner, err = sim.NewRunner(*uri, *tlsCAFile, *tlsCertificateKeyFile); err != nil {
		log.Fatal(err)
	}
	runner.SetTPS(*tps)
	runner.SetTemplateFilename(*file)
	runner.SetVerbose(*verbose)
	runner.SetSimulationDuration(*duration)
	runner.SetPeekingMode(*peek)
	runner.SetDropFirstMode(*drop)
	runner.SetNumberConnections(*conn)
	runner.SetTransactionTemplateFilename(*tx)
	runner.SetSimOnlyMode(*simonly)
	runner.SetAutoMode(*yes)
	if err = runner.Start(); err != nil {
		log.Fatal(err)
	}
	runner.CollectAllStatus()
}

func handler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": 1, "message": "hello keyhole!"})
}
