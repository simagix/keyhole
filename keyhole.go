// Copyright 2020-present Kuei-chun Chen. All rights reserved.
// keyhole.go

package keyhole

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/simagix/gox"
	"github.com/simagix/keyhole/atlas"
	"github.com/simagix/keyhole/mdb"
	"github.com/simagix/keyhole/sim"
	"github.com/simagix/keyhole/sim/util"
	ftdc "github.com/simagix/mongo-ftdc"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

const (
	outdir  = "./out"
	htmldir = "./html"
)

// Run executes main()
func Run(fullVersion string) {
	var err error
	var dbNames IncludeDB

	config := flag.String("config", "", "configuration file")

	allinfo := flag.String("allinfo", "", "database connection string, used with optional -db")
	cardinality := flag.String("cardinality", "", "check collection cardinality")
	changeStreams := flag.Bool("changeStreams", false, "change streams watch")
	collection := flag.String("collection", "", "collection name to print schema")
	collscan := flag.Bool("collscan", false, "list only COLLSCAN (with --loginfo)")
	compare := flag.Bool("compare", false, "(deprecated) compare 2 clusters or 2 -allinfo output files")
	conn := flag.Int("conn", 0, "number of connections")
	createIndex := flag.String("createIndex", "", "create indexes")
	flag.Var(&dbNames, "db", `database to include with -allinfo`)
	diag := flag.String("diag", "", "diagnosis of server status or diagnostic.data")
	drop := flag.Bool("drop", false, "drop examples collection before seeding")
	duration := flag.Int("duration", 5, "load test duration in minutes")
	explain := flag.String("explain", "", "explain a query from a JSON doc or a log line")
	file := flag.String("file", "", "template file for seedibg data")
	ftdcOn := flag.Bool("ftdc", false, "download from atlas://user:key@group/cluster")
	index := flag.String("index", "", "get indexes info")
	info := flag.String("info", "", "database connection string (Atlas uses atlas://user:key)")
	loginfo := flag.Bool("loginfo", false, "log performance analytic from file or Atlas")
	maobiURL := flag.String("maobi", "", "maobi url")
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
	tps := flag.Int("tps", 20, "number of trasaction per second per connection")
	total := flag.Int("total", 1000, "number of documents to create")
	tx := flag.String("tx", "", "file with defined transactions")
	ver := flag.Bool("version", false, "print version number")
	verbose := flag.Bool("v", false, "verbose")
	viewlog := flag.String("viewlog", "", "view v4.4+ log file")
	webserver := flag.Bool("web", false, "enable web server")
	wt := flag.Bool("wt", false, "visualize wiredTiger cache usage")
	yes := flag.Bool("yes", false, "bypass confirmation")

	flag.Parse()
	flagset := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { flagset[f.Name] = true })

	if *config != "" {
		if err = Exec(*config, fullVersion); err != nil {
			log.Fatal(err)
		}
		return
	}

	var uri string
	if *allinfo != "" {
		uri = *allinfo
	} else if *info != "" {
		uri = *info
	} else if *index != "" {
		uri = *index
	} else if len(flag.Args()) > 0 {
		uri = flag.Arg(0)
	}
	if *maobiURL == "" && os.Getenv("MAOBI") != "" {
		*maobiURL = os.Getenv("MAOBI")
	}

	if strings.HasPrefix(uri, "atlas://") {
		var api *atlas.API
		if api, err = atlas.ParseURI(uri); err != nil {
			log.Fatal(err)
		}
		api.SetArgs(flag.Args())
		api.SetFTDC(*ftdcOn)
		if *info != "" {
			api.SetInfo(true)
		}
		api.SetLoginfo(*loginfo)
		api.SetPause(*pause)
		api.SetResume(*resume)
		api.SetRequest(*request)
		api.SetVerbose(*verbose)
		fmt.Println(api.Execute())

		if *loginfo {
			l := mdb.NewLogInfo(fullVersion)
			l.SetCollscan(*collscan)
			l.SetRedaction(*redaction)
			l.SetRegexPattern(*regex)
			l.SetSilent(*nocolor)
			l.SetVerbose(*verbose)
			if err = AnalyzeMongoLogs(l, api.GetLogNames(), *maobiURL); err != nil {
				log.Fatal(err)
			}
		}
		return
	} else if *compare {
		PrintCompareHelp()
		return
	} else if *webserver {
		addr := fmt.Sprintf(":%d", *port)
		if listener, err := net.Listen("tcp", addr); err != nil {
			log.Fatal(err)
		} else {
			listener.Close()
		}
		filenames := append([]string{*diag}, flag.Args()...)
		metrics := ftdc.NewMetrics()
		if err = metrics.ProcessFiles(filenames); err != nil {
			log.Fatal(err)
		}
		log.Fatal(http.ListenAndServe(addr, nil))
	} else if *diag != "" {
		filenames := append([]string{*diag}, flag.Args()...)
		metrics := ftdc.NewDiagnosticData()
		if err := metrics.DecodeDiagnosticData(filenames); err != nil {
			log.Fatal(err)
		}
		strs := []string{}
		if metrics.ServerInfo != nil {
			var p mdb.ClusterStats
			b, _ := json.Marshal(metrics.ServerInfo)
			json.Unmarshal(b, &p)

			result := fmt.Sprintf(`MongoDB v%v %v (%v) %v %v %v cores %v mem`,
				p.BuildInfo.Version, p.HostInfo.System.Hostname, p.HostInfo.OS.Name,
				p.ServerStatus.Process, p.Cluster, p.HostInfo.System.NumCores, p.HostInfo.System.MemSizeMB)
			strs = append(strs, result)
		}
		strs = append(strs, ftdc.PrintAllStats(metrics.ServerStatusList, -1))
		fmt.Println(strings.Join(strs, "\n"))
		return
	} else if *ver {
		fmt.Println(fullVersion)
		return
	} else if *explain != "" && uri == "" { //--explain file.json.gz (w/o uri)
		exp := mdb.NewExplain()
		if err = exp.PrintExplainResults(*explain); err != nil {
			log.Fatal(err)
		}
		return
	} else if *loginfo && len(flag.Args()) > 0 {
		l := mdb.NewLogInfo(fullVersion)
		l.SetCollscan(*collscan)
		l.SetRedaction(*redaction)
		l.SetRegexPattern(*regex)
		l.SetSilent(*nocolor)
		l.SetVerbose(*verbose)
		if err = AnalyzeMongoLogs(l, flag.Args(), *maobiURL); err != nil {
			log.Fatal(err)
		}
		return
	} else if *print != "" {
		printer := mdb.NewBSONPrinter(fullVersion)
		printer.SetNoColor(*nocolor)
		if err := printer.Print(*print); err != nil {
			log.Fatal(err)
		}
		return
	} else if *viewlog != "" {
		mdb.OutputLogInOldFormat(*viewlog)
		return
	} else if uri == "" {
		flag.PrintDefaults()
		fmt.Println("\nusage: keyhole [options] <connection_string>")
		return
	}

	var client *mongo.Client
	tag := "readPreferenceTags=nodeType:ANALYTICS"
	var fastMode bool
	if strings.Contains(uri, ".mongodb.net") && strings.Contains(uri, tag) {
		fastMode = true
		if *allinfo != "" {
			gox.GetLogger(fullVersion).Infof(`remove "%v" from connection string for fast processing`, tag)
		}
		uri = strings.Replace(uri, tag, "", -1)
	}
	// connection string is required from here forward
	var connString connstring.ConnString
	if connString, err = mdb.ParseURI(uri); err != nil {
		log.Fatal(err)
	}
	uri = connString.String() // password can be injected if missing
	if client, err = mdb.NewMongoClient(uri); err != nil {
		log.Fatal(err)
	}

	if *allinfo != "" {
		var data []byte
		var ofile string
		stats := mdb.NewClusterStats(fullVersion)
		stats.SetDBNames(dbNames)
		stats.SetRedaction(*redaction)
		stats.SetVerbose(*verbose)
		stats.SetFastMode(fastMode)
		if err = stats.GetClusterStats(client, connString); err != nil {
			log.Fatalf("a valid user with roles 'clusterMonitor' and 'readAnyDatabase' on all mongo processes are required.\n%v", err)
		}
		stats.Print()
		if ofile, data, err = stats.OutputBSON(); err != nil {
			log.Fatalf("failed to output bson file\n%v", err)
		}
		GenerateMaobiReport(*maobiURL, data, ofile)
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
	} else if *changeStreams {
		stream := mdb.NewChangeStream()
		stream.SetCollection(*collection)
		stream.SetDatabase(connString.Database)
		stream.SetPipelineString(*pipe)
		stream.Watch(client, util.Echo)
		return
	} else if *createIndex != "" {
		ix := mdb.NewIndexStats(fullVersion)
		ix.SetNoColor(*nocolor)
		ix.SetVerbose(*verbose)
		ix.SetFastMode(fastMode)
		if err = DuplicateIndexesFromFile(ix, client, *createIndex, *drop); err != nil {
			log.Fatal(err)
		}
		return
	} else if *explain != "" { // --explain json_or_log_file  [-v]
		exp := mdb.NewExplain()
		exp.SetVerbose(*verbose)
		if err = exp.ExecuteAllPlans(client, *explain); err != nil {
			log.Fatal(err)
		}
		return
	} else if *index != "" {
		ix := mdb.NewIndexStats(fullVersion)
		ix.SetNoColor(*nocolor)
		ix.SetVerbose(*verbose)
		ix.SetFastMode(fastMode)
		if err = CollectIndexStats(ix, client, *maobiURL); err != nil {
			log.Fatal(err)
		}
		return
	} else if *schema {
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
	} else if *seed {
		f := NewSeed()
		f.SetCollection(*collection)
		f.SetDatabase(connString.Database)
		f.SetFile(*file)
		f.SetIsDrop(*drop)
		f.SetNumberConnections(*conn)
		f.SetTotal(*total)
		if err = f.SeedData(client); err != nil {
			log.Fatal(err)
		}
		return
	}

	clusterSummary := GetClusterSummary(fullVersion, client)
	if *info != "" {
		fmt.Println(clusterSummary)
		return
	}

	go func() {
		http.HandleFunc("/", gox.Cors(handler))
		addr := fmt.Sprintf(":%d", *port)
		log.Println(http.ListenAndServe(addr, nil))
	}()
	if *wt {
		fmt.Println(clusterSummary)
		log.Printf("URL: http://localhost:%d/wt\n", *port)
		MonitorWiredTigerCache(fullVersion, client)
	}

	var runner *sim.Runner
	if runner, err = sim.NewRunner(connString); err != nil {
		log.Fatal(err)
	}
	logger := gox.GetLogger(fullVersion)
	logger.Info(fullVersion)
	logger.Info(clusterSummary)
	runner.SetAutoMode(*yes)
	runner.SetCollection(*collection)
	runner.SetDropFirstMode(*drop)
	runner.SetDuration(*duration)
	runner.SetNumberConnections(*conn)
	runner.SetPeekingMode(*peek)
	runner.SetSimOnlyMode(*simonly)
	runner.SetTemplateFilename(*file)
	runner.SetTPS(*tps)
	runner.SetTransactionTemplate(*tx)
	runner.SetVerbose(*verbose)

	if err = StartSimulation(runner); err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": 1, "message": "hello keyhole!"})
}

// GetClusterSummary returns MongoDB cluster summary
func GetClusterSummary(version string, client *mongo.Client) string {
	stats := mdb.NewClusterStats(version)
	return stats.GetClusterShortSummary(client)
}
