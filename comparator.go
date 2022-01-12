// Copyright 2020-present Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"context"
	"fmt"
	"html/template"
	"os"
	"reflect"
	"runtime"
	"sort"
	"sync"

	"github.com/simagix/gox"
	"github.com/simagix/keyhole/mdb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// Comparator stores validation info
type Comparator struct {
	messages []ErrorMessage
	mutex    sync.Mutex
	nsMap    map[string]NamespaceStats
	source   *mongo.Client
	target   *mongo.Client
}

// NamespaceStats stores migration counts
type NamespaceStats struct {
	NS            string `json:"ns" bson:"ns"`
	Query         bson.D `json:"query" bson:"query"`
	SourceCount   int64  `json:"source_counts" bson:"source_counts"`
	SourceIndexes int64  `json:"source_indexes" bson:"source_indexes"`
	TargetCount   int64  `json:"target_counts" bson:"target_counts"`
	TargetIndexes int64  `json:"target_indexes" bson:"target_indexes"`
}

// ComparatorStats stores data for HTML
type ComparatorStats struct {
	Messages []ErrorMessage
	Status   []NamespaceStats
}

// ErrorMessage keeps error message and info
type ErrorMessage struct {
	Descr string
	Error string
	NS    string
}

// NewComparator returns Comparator
func NewComparator(source string, target string) (*Comparator, error) {
	var err error
	var logger = gox.GetLogger("comparator")
	var validator = Comparator{nsMap: map[string]NamespaceStats{}, messages: []ErrorMessage{},
		mutex: sync.Mutex{}}
	if validator.source, err = GetMongoClientFromPool(source); err != nil {
		logger.Errorf(`error connecting %v`, source)
		return &validator, err
	}
	stats := &mdb.ClusterStats{Logger: logger}
	if err = stats.GetClusterStatsSummary(validator.source); err != nil {
		return &validator, err
	}
	logger.Infof(`source: %v`, stats.GetClusterShortSummary(validator.source))
	if validator.target, err = GetMongoClientFromPool(target); err != nil {
		logger.Errorf(`error connecting %v`, target)
		return &validator, err
	}
	if err = stats.GetClusterStatsSummary(validator.target); err != nil {
		return &validator, err
	}
	logger.Infof(`target: %v`, stats.GetClusterShortSummary(validator.target))
	return &validator, err
}

// Compare Compare namespaces
func (p *Comparator) Compare(filters []Filter, sampleSize int) error {
	var err error
	var namespaces []string
	var logger = gox.GetLogger("comparator")

	if len(filters) == 0 {
		if namespaces, err = GetQualifiedNamespaces(p.source); err != nil {
			return err
		}
		var list []Filter
		for _, ns := range namespaces {
			list = append(list, Filter{NS: ns, Query: bson.D{}, TargetNS: ns})
		}
		filters = list
	} else {
		var list []Filter
		for _, filter := range filters {
			dbName, collName := mdb.SplitNamespace(filter.NS)
			if collName == "" {
				if namespaces, err = GetQualifiedNamespacesByDB(p.source, dbName); err != nil {
					continue
				}
				for _, ns := range namespaces {
					list = append(list, Filter{NS: ns, Query: bson.D{}, TargetNS: ns})
				}
				continue
			}
			if filter.TargetNS == "" {
				filter.TargetNS = filter.NS
			}
			list = append(list, filter)
		}
		filters = list
	}
	sort.Slice(filters, func(i int, j int) bool {
		return filters[i].NS < filters[j].NS
	})

	var wg = gox.NewWaitGroup(runtime.NumCPU())
	logger.Info("compare document counts")
	for _, filter := range filters {
		p.mutex.Lock()
		p.nsMap[filter.NS] = NamespaceStats{NS: filter.NS, Query: filter.Query}
		p.mutex.Unlock()
		logger.Infof(" - count %v using filter %v", filter.NS, Stringify(filter.Query))
		// count source
		wg.Add(1)
		go func(filter Filter) {
			defer wg.Done()
			ctx := context.Background()
			var counts int64
			dbName, collName := mdb.SplitNamespace(filter.NS)
			opts := options.Count()
			if len(filter.Query) == 0 {
				opts.SetHint(bson.D{{Key: "_id", Value: 1}})
			}
			if counts, err = p.source.Database(dbName).Collection(collName).CountDocuments(ctx, filter.Query, opts); err != nil {
				logger.Error("source", err)
				return
			}
			p.setSourceCounts(filter.NS, counts)
		}(filter)
		// count target
		wg.Add(1)
		go func(filter Filter) {
			defer wg.Done()
			ctx := context.Background()
			var counts int64
			dbName, collName := mdb.SplitNamespace(filter.TargetNS)
			opts := options.Count()
			if len(filter.Query) == 0 {
				opts.SetHint(bson.D{{Key: "_id", Value: 1}})
			}
			if counts, err = p.target.Database(dbName).Collection(collName).CountDocuments(ctx, filter.Query, opts); err != nil {
				logger.Error("target", err)
				return
			}
			p.setTargetCounts(filter.NS, counts)
		}(filter)
	}
	wg.Wait()

	logger.Info("compare indexes")
	idx := mdb.NewIndexStats("")
	for _, filter := range filters {
		dbName, collName := mdb.SplitNamespace(filter.NS)
		var sourceIndexes []mdb.Index
		if sourceIndexes, err = idx.GetIndexesFromCollection(p.source, p.source.Database(dbName).Collection(collName)); err != nil {
			logger.Error("source: ", err)
			continue
		}
		dbName, collName = mdb.SplitNamespace(filter.TargetNS)
		var targetIndexes []mdb.Index
		if targetIndexes, err = idx.GetIndexesFromCollection(p.target, p.target.Database(dbName).Collection(collName)); err != nil {
			logger.Error("source: ", err)
		}
		targetIndexMap := map[string]*mdb.Index{}
		for i, value := range targetIndexes {
			targetIndexMap[value.Name] = &targetIndexes[i]
		}
		p.setSourceIndexs(filter.NS, len(sourceIndexes))
		p.setTargetIndexs(filter.NS, len(targetIndexes))
		if len(sourceIndexes) != len(targetIndexes) {
			logger.Errorf(` - %v number of indexes not the same, source %v  != target %v`, filter.NS, len(sourceIndexes), len(targetIndexes))
		} else {
			logger.Infof(` - %v number of indexes checked (%v)`, filter.NS, len(sourceIndexes))
		}
		for _, value := range sourceIndexes {
			if targetIndexMap[value.Name] == nil {
				message := ErrorMessage{NS: filter.NS, Error: "missing", Descr: fmt.Sprintf(`index: "%v"`, value.Name)}
				logger.Errorf(" - %v", message)
				p.messages = append(p.messages, message)
			}
		}
	}
	logger.Info("compare documents")
	if err = p.spotCheck(filters, sampleSize); err != nil {
		return err
	}
	pr := message.NewPrinter(language.English)
	fmt.Println()
	pr.Printf("%4s%-40s%14s%14s   %8s%8s\n", " ", "namespace", "source", "target", "source", "target")
	pr.Printf("%4s%-40s%14s%14s   %8s%8s\n", " ", " ", " count", " count", " index", " index")
	pr.Println("===========================================================================================")
	var docs []NamespaceStats
	for i, filter := range filters {
		s1 := " "
		s2 := " "
		value := p.getMigrationStatus(filter.NS)
		if value.SourceCount != value.TargetCount {
			s1 = "!"
		}
		if value.SourceIndexes != value.TargetIndexes {
			s2 = "!"
		}
		pr.Printf("%3d %-40s%14d%14d %s %8d%8d %s\n", i+1, filter.NS, value.SourceCount,
			value.TargetCount, s1, value.SourceIndexes, value.TargetIndexes, s2)
		docs = append(docs, value)
	}
	fmt.Println()
	WirteValidatorHTML(ComparatorStats{Messages: p.messages, Status: docs})
	return err
}

func (p *Comparator) spotCheck(filters []Filter, sampleSize int) error {
	var wg = gox.NewWaitGroup(runtime.NumCPU())
	for _, filter := range filters {
		wg.Add(1)
		go func(filter Filter) error {
			defer wg.Done()
			var err error
			var cursor *mongo.Cursor
			ctx := context.Background()
			var ids []interface{}
			dbName, collName := mdb.SplitNamespace(filter.NS)
			sample := bson.D{{Key: "$sample", Value: bson.D{{Key: "size", Value: sampleSize}}}}
			var pipeline mongo.Pipeline
			if len(filter.Query) > 0 {
				pipeline = mongo.Pipeline{bson.D{{Key: "$match", Value: filter.Query}}}
			}
			if sampleSize > 0 {
				pipeline = append(pipeline, sample)
			}
			if len(pipeline) == 0 {
				pipeline = mongo.Pipeline{}
			}
			opts := options.Aggregate()
			opts.SetAllowDiskUse(true)
			if cursor, err = p.source.Database(dbName).Collection(collName).Aggregate(ctx, pipeline, opts); err != nil {
				return err
			}
			docsMap := map[interface{}]bson.D{}
			sourceCount := 0
			memsize := 0
			for cursor.Next(ctx) {
				memsize += len(cursor.Current)
				if memsize > 64*mb {
					messages := compareTarget(p.target, filter, ids, docsMap)
					p.messages = append(p.messages, messages...)
					sourceCount = 0
					docsMap = map[interface{}]bson.D{}
					ids = nil
					memsize = len(cursor.Current)
				}
				var doc bson.D
				cursor.Decode(&doc)
				sourceCount++
				ids = append(ids, doc.Map()["_id"])
				docsMap[doc.Map()["_id"]] = doc
			}
			cursor.Close(ctx)
			messages := compareTarget(p.target, filter, ids, docsMap)
			p.messages = append(p.messages, messages...)
			return nil
		}(filter)
	}
	wg.Wait()
	return nil
}

func compareTarget(client *mongo.Client, filter Filter, ids []interface{}, docsMap map[interface{}]bson.D) []ErrorMessage {
	var err error
	var messages []ErrorMessage
	var cursor *mongo.Cursor
	ctx := context.Background()
	targetCount := 0
	logger := gox.GetLogger("comparator")
	pr := message.NewPrinter(language.English)

	dbName, collName := mdb.SplitNamespace(filter.TargetNS)
	query := bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: ids}}}}
	if len(ids) == 0 {
		query = bson.D{}
	}
	if cursor, err = client.Database(dbName).Collection(collName).Find(ctx, query); err != nil {
		return messages
	}
	for cursor.Next(ctx) {
		var doc bson.D
		cursor.Decode(&doc)
		if !reflect.DeepEqual(doc, docsMap[doc.Map()["_id"]]) {
			message := ErrorMessage{NS: filter.NS, Error: "diff", Descr: fmt.Sprintf(`{"_id": %v}`, doc.Map()["_id"])}
			logger.Error(message)
			messages = append(messages, message)
		} else {
			targetCount++
		}
		delete(docsMap, doc.Map()["_id"])
	}
	sourceCount := len(ids)
	if sourceCount != targetCount {
		// message := fmt.Sprintf(`"%v" failed to get equal number of docs validated, source: %v, target: %v`,
		// 	ns, pr.Sprintf("%d", sourceCount), pr.Sprintf("%d", targetCount))
		// logger.Warn(message)
		for k := range docsMap {
			val := k
			switch k.(type) {
			case string:
				val = fmt.Sprintf(`"%v"`, k)
			}
			message := ErrorMessage{NS: filter.NS, Error: "missing", Descr: fmt.Sprintf(`{"_id": %v}`, val)}
			logger.Errorf(` - %v missing doc {"_id": %v}`, filter.TargetNS, val)
			messages = append(messages, message)
		}
	} else {
		message := fmt.Sprintf(`"%v" has %v docs validated`, filter.NS, pr.Sprintf("%d", sourceCount))
		logger.Info(message)
	}
	return messages
}

func (p *Comparator) getMigrationStatus(ns string) NamespaceStats {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.nsMap[ns]
}

func (p *Comparator) setSourceCounts(ns string, counts int64) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	m := p.nsMap[ns]
	m.SourceCount = counts
	p.nsMap[ns] = m
}

func (p *Comparator) setTargetCounts(ns string, counts int64) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	m := p.nsMap[ns]
	m.TargetCount = counts
	p.nsMap[ns] = m
}

func (p *Comparator) setSourceIndexs(ns string, counts int) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	m := p.nsMap[ns]
	m.SourceIndexes = int64(counts)
	p.nsMap[ns] = m
}

func (p *Comparator) setTargetIndexs(ns string, counts int) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	m := p.nsMap[ns]
	m.TargetIndexes = int64(counts)
	p.nsMap[ns] = m
}

// WirteValidatorHTML outputs validators to a HTML file
func WirteValidatorHTML(stats ComparatorStats) error {
	var err error
	os.Mkdir(htmldir, 0755)
	ofile := fmt.Sprintf(`%v/keyhole-compare.html`, htmldir)
	i := 1
	for mdb.DoesFileExist(ofile) {
		ofile = fmt.Sprintf(`%v/keyhole.%d-compare.html`, htmldir, i)
		i++
	}
	var w *os.File
	if w, err = os.Create(ofile); err != nil {
		return err
	}
	templ, err := GetValidatorTemplate()
	templ.Execute(w, stats)
	fmt.Println("html report written to", ofile)
	return err
}

// GetValidatorTemplate returns HTML of -compare
func GetValidatorTemplate() (*template.Template, error) {
	return template.New("validator").Funcs(template.FuncMap{
		"printTarget": func(target interface{}, source interface{}) template.HTML {
			printer := message.NewPrinter(language.English)
			if target == source {
				return template.HTML(printer.Sprintf(`%v`, target))
			}
			return template.HTML(fmt.Sprintf(`<span style="color:red; background-color: #FFFF00">%v</span>`, printer.Sprintf("%v", target)))
		},
		"numPrinter": func(n int64) string {
			printer := message.NewPrinter(language.English)
			return printer.Sprintf("%v", n)
		},
		"add": func(a int, b int) string {
			return fmt.Sprintf("%v", a+b)
		},
		"IsPrint": func(docs []ErrorMessage) bool {
			return (len(docs) > 0)
		},
	}).Parse(validatorTemplate)
}

const validatorTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <title>Ken Chen's Keyhole Clusters Comparator</title>
  <style>
  	body {
		font-family: Helvetica, Arial, sans-serif;
		background-color: #f2f2f2;
		margin-top: 30px;
		margin-bottom: 30px;
		margin-right: 150px;
		margin-left: 150px;
  	}
    table
    {
    	font-family: Consolas, monaco, monospace;
    	border-collapse:collapse;
    	min-width:600px;
    }
    caption
    {
    	caption-side:top;
    	font-weight:bold;
    	font-style:italic;
    	margin:2px;
    }
    table, th, td
    {
		border: 1px solid gray;
		vertical-align: top;
    }
    th, td
    {
    	padding:2px;
    	vertical-align: top;
    }
    th
    {
      background-color: #a0c3ff;
      color: black;
    	font-weight:bold;
    }
    tr:nth-child(odd) {background-color: #f2f2f2;}
    tr:nth-child(even) {background-color: #fff;}
    .rowtitle
    {
    	font-weight:bold;
    }
    h1 {
			font-family: "Trebuchet MS";
      font-size: 1.7em;
      font-weight: bold;
    }
    h2 {
			font-family: "Trebuchet MS";
      font-size: 1.5em;
      font-weight: bold;
    }
    h3 {
			font-family: "Trebuchet MS";
      font-size: 1.25em;
      font-weight: bold;
    }
    </style>
</head>

<body>
<h1>Keyhole Clusters Comparator</h1>
<table>
	<tr><th></th><th>Namespace</th><th>Source Counts</th><th>Target Counts</th><th>Source Indexes</th><th>Target Indexes</th></tr>
	{{range $k, $v := .Status}}
			<tr>
				<td align='right'>{{add $k 1}}</td>
				<td align='left'>{{$v.NS}}</td>
				<td align='right'>{{numPrinter $v.SourceCount}}</td>
			    <td align='right'>{{printTarget $v.TargetCount $v.SourceCount}}</td>
				<td align='right'>{{numPrinter $v.SourceIndexes}}</td> 
			    <td align='right'>{{printTarget $v.TargetIndexes $v.SourceIndexes}}</td>
			</tr>
	{{end}}
</table>
{{ if (IsPrint .Messages) }}
<hr/>
<table>
	<tr><th></th><th>Namespace</th><th>Error</th><th>Description</th></tr>
	{{range $i, $v := .Messages}}
			<tr>
				<td align='right'>{{add $i 1}}</td>
				<td align='left'>{{$v.NS}}</td>
				<td align='left'>{{$v.Error}}</td>
				<td align='left'>{{$v.Descr}}</td>
			</tr>
	{{end}}
</table>
{{ end }}
</body>
`
