// Copyright 2020 Kuei-chun Chen. All rights reserved.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/simagix/gox"
	"github.com/simagix/keyhole/mdb"
	"github.com/simagix/keyhole/sim"
	"github.com/simagix/keyhole/sim/util"
	"github.com/simagix/mongo-atlas/atlas"
	anly "github.com/simagix/mongo-ftdc/analytics"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

var version = "devel-xxxxxx"

func main() {
	var err error
	if version == "devel-xxxxxx" {
		version = "devel-" + time.Now().Format("20060102")
	}

	allinfo := flag.Bool("allinfo", false, "get all cluster info")
	cardinality := flag.String("cardinality", "", "check collection cardinality")
	changeStreams := flag.Bool("changeStreams", false, "change streams watch")
	collection := flag.String("collection", "", "collection name to print schema")
	collscan := flag.Bool("collscan", false, "list only COLLSCAN (with --loginfo)")
	conn := flag.Int("conn", 0, "nuumber of connections")
	createIndex := flag.String("createIndex", "", "create indexes")
	diag := flag.String("diag", "", "diagnosis of server status or diagnostic.data")
	drop := flag.Bool("drop", false, "drop examples collection before seeding")
	duration := flag.Int("duration", 5, "load test duration in minutes")
	explain := flag.String("explain", "", "explain a query from a JSON doc or a log line")
	file := flag.String("file", "", "template file for seedibg data")
	ftdc := flag.Bool("ftdc", false, "download from atlas://user:key@group/cluster")
	index := flag.Bool("index", false, "get indexes info")
	info := flag.Bool("info", false, "Atlas info (atlas://user:key)")
	loginfo := flag.Bool("loginfo", false, "log performance analytic from file or Atlas")
	nocolor := flag.Bool("nocolor", false, "disable color codes")
	pause := flag.Bool("pause", false, "pause an Atlas cluster atlas://user:key@group/cluster")
	peek := flag.Bool("peek", false, "only collect stats")
	pipe := flag.String("pipeline", "", "aggregation pipeline")
	port := flag.Int("port", 5408, "web server port number")
	print := flag.String("print", "", "print contents of input file")
	redaction := flag.Bool("redact", false, "redact document")
	regex := flag.String("regex", "", "regex pattern for loginfo")
	request := flag.String("request", "", "Atlas API command")
	resume := flag.Bool("resume", false, "resume an Atlas cluster atlas://user:key@group/cluster")
	schema := flag.Bool("schema", false, "print schema")
	seed := flag.Bool("seed", false, "seed a database for demo")
	simonly := flag.Bool("simonly", false, "simulation only mode")
	tlsCAFile := flag.String("tlsCAFile", "", "TLS CA file")
	tlsCertificateKeyFile := flag.String("tlsCertificateKeyFile", "", "TLS CertificateKey File")
	tps := flag.Int("tps", 20, "number of trasaction per second per connection")
	total := flag.Int("total", 1000, "nuumber of documents to create")
	tx := flag.String("tx", "", "file with defined transactions")
	ver := flag.Bool("version", false, "print version number")
	verbose := flag.Bool("v", false, "verbose")
	webserver := flag.Bool("web", false, "enable web server")
	wt := flag.Bool("wt", false, "visualize wiredTiger cache usage")
	yes := flag.Bool("yes", false, "bypass confirmation")

	flag.Parse()
	flagset := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { flagset[f.Name] = true })
	var uri string
	if len(flag.Args()) > 0 {
		uri = flag.Arg(0)
	}

	if strings.HasPrefix(uri, "atlas://") {
		var api *atlas.API
		if api, err = atlas.ParseURI(uri); err != nil {
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
				li := mdb.NewLogInfo(version)
				li.SetVerbose(*verbose)
				if str, err = li.AnalyzeFile(filename, *redaction); err != nil {
					log.Println(err)
					continue
				}
				fmt.Println(str)
				if li.OutputFilename != "" {
					log.Println("Log info written to", li.OutputFilename)
				}
			}
		}
		return
	} else if *webserver {
		filenames := append([]string{*diag}, flag.Args()...)
		addr := fmt.Sprintf(":%d", *port)
		if listener, err := net.Listen("tcp", addr); err != nil {
			log.Fatal(err)
		} else {
			listener.Close()
		}
		metrics := anly.NewMetrics()
		metrics.ProcessFiles(filenames)
		log.Fatal(http.ListenAndServe(addr, nil))
	} else if *diag != "" {
		filenames := append([]string{*diag}, flag.Args()...)
		metrics := anly.NewDiagnosticData()
		if str, e := metrics.PrintDiagnosticData(filenames); e != nil {
			log.Fatal(e)
		} else {
			fmt.Println(str)
		}
		return
	} else if *ver {
		fmt.Println("keyhole", version)
		return
	} else if *explain != "" && uri == "" { //--explain file.json.gz (w/o uri)
		exp := mdb.NewExplain()
		if err = exp.PrintExplainResults(*explain); err != nil {
			log.Fatal(err)
		}
		return
	} else if *loginfo && len(flag.Args()) > 0 {
		filenames := flag.Args()
		li := mdb.NewLogInfo(version)
		li.SetRegexPattern(*regex)
		li.SetCollscan(*collscan)
		li.SetVerbose(*verbose)
		li.SetSilent(*nocolor)
		for _, filename := range filenames {
			var str string
			if str, err = li.AnalyzeFile(filename, *redaction); err != nil {
				log.Fatal(err)
			}
			fmt.Println(str)
			if li.OutputFilename != "" {
				log.Println("Log info written to", li.OutputFilename)
			}
		}
		return
	} else if *print != "" {
		printer := mdb.NewBSONPrinter(version)
		if err := printer.Translate(*print); err != nil {
			log.Fatal(err)
		}
		return
	} else if len(flag.Args()) == 0 {
		flag.PrintDefaults()
		fmt.Println("\nusage: keyhole [options] <connection_string>")
		return
	}

	var client *mongo.Client
	// connection string is required from here forward
	var connString connstring.ConnString
	if connString, err = mdb.ParseURI(uri); err != nil {
		log.Fatal(err)
	}
	uri = connString.String() // password can be injected if missing
	connString.SSLCaFile = *tlsCAFile
	connString.SSLClientCertificateKeyFile = *tlsCertificateKeyFile
	if client, err = mdb.NewMongoClient(uri, *tlsCAFile, *tlsCertificateKeyFile); err != nil {
		log.Fatal(err)
	}

	var result string
	if *info && *verbose {
		*allinfo = true
	}

	if *allinfo {
		keyhole := mdb.NewKeyhole(version)
		if result, err = keyhole.CollectClusterStats(client, connString); err != nil {
			log.Fatal(err)
		}
		fmt.Println(result)
		return
	} else if *cardinality != "" { // --card <collection> [-v]
		card := mdb.NewCardinality(client)
		card.SetVerbose(*verbose)
		if summary, e := card.GetCardinalityArray(connString.Database, *cardinality); e != nil {
			log.Fatal(e)
		} else {
			fmt.Println(card.GetSummary(summary))
		}
		return
	} else if *changeStreams == true {
		stream := mdb.NewChangeStream()
		stream.SetCollection(*collection)
		stream.SetDatabase(connString.Database)
		stream.SetPipelineString(*pipe)
		stream.Watch(client, util.Echo)
		return
	} else if *createIndex != "" {
		ix := mdb.NewIndexStats(version)
		ix.SetNoColor(*nocolor)
		ix.SetVerbose(*verbose)
		if err = ix.SetIndexesMapFromFile(*createIndex); err != nil {
			log.Fatal(err)
		}
		if err = ix.CreateIndexes(client); err != nil {
			log.Fatal(err)
		}
		if indexesMap, ixe := ix.GetIndexes(client); ixe != nil {
			log.Fatal(err)
		} else {
			ix.PrintIndexesOf(indexesMap)
		}
		return
	} else if *explain != "" { // --explain json_or_log_file  [-v]
		exp := mdb.NewExplain()
		exp.SetVerbose(*verbose)
		if err = exp.ExecuteAllPlans(client, *explain); err != nil {
			log.Fatal(err)
		}
		return
	} else if *index == true {
		ix := mdb.NewIndexStats(version)
		ix.SetNoColor(*nocolor)
		if connString.Database == mdb.KeyholeDB {
			connString.Database = ""
		}
		ix.SetDBName(connString.Database)
		ix.SetVerbose(*verbose)
		if indexesMap, ixe := ix.GetIndexes(client); ixe != nil {
			log.Fatal(err)
		} else {
			ix.PrintIndexesOf(indexesMap)
			if err = ix.Save(); err != nil {
				log.Fatal(err)
			}
		}
		return
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
		return
	} else if *seed == true {
		f := sim.NewFeeder()
		f.SetCollection(*collection)
		f.SetDatabase(connString.Database)
		f.SetFile(*file)
		f.SetIsDrop(*drop)
		nConnection := 2 * runtime.NumCPU()
		if *conn != 0 {
			nConnection = *conn
		}
		f.SetNumberConnections(nConnection)
		f.SetTotal(*total)
		if err = f.SeedData(client); err != nil {
			log.Fatal(err)
		}
		return
	}

	keyhole := mdb.NewKeyhole(version)
	fmt.Println(keyhole.GetClusterSummary(client))
	if *info == true {
		return
	}

	go func() {
		http.HandleFunc("/", gox.Cors(handler))
		addr := fmt.Sprintf(":%d", *port)
		log.Println(http.ListenAndServe(addr, nil))
	}()
	if *wt == true {
		wtc := mdb.NewWiredTigerCache(version)
		log.Printf("URL: http://localhost:%d/wt\n", *port)
		wtc.Start(client)
	}

	var runner *sim.Runner
	if runner, err = sim.NewRunner(connString); err != nil {
		log.Fatal(err)
	}
	runner.SetCollection(*collection)
	runner.SetTPS(*tps)
	runner.SetTemplateFilename(*file)
	runner.SetVerbose(*verbose)
	runner.SetSimulationDuration(*duration)
	runner.SetPeekingMode(*peek)
	runner.SetDropFirstMode(*drop)
	nConnection := runtime.NumCPU()
	if *conn != 0 {
		nConnection = *conn
	}
	runner.SetNumberConnections(nConnection)
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
