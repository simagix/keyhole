// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// IndexesReader holder indexes reader struct
type IndexesReader struct {
	client  *mongo.Client
	dbName  string
	nocolor bool
	verbose bool
}

// AccessesDoc - accessss
type AccessesDoc struct {
	Ops   int       `json:"ops" bson:"ops"`
	Since time.Time `json:"since" bson:"since"`
}

// UsageDoc -
type UsageDoc struct {
	Host     string      `json:"host"`
	Accesses AccessesDoc `json:"accesses"`
}

// IndexStatsDoc -
type IndexStatsDoc struct {
	Fields       []string
	Key          string     `json:"key"`
	Name         string     `json:"name"`
	EffectiveKey string     `json:"effectiveKey"`
	IsDupped     bool       `json:"dupped"`
	IsShardKey   bool       `json:"shardKey"`
	TotalOps     int        `json:"totalOps"`
	Usage        []UsageDoc `json:"stats"`
}

// NewIndexesReader establish seeding parameters
func NewIndexesReader(client *mongo.Client) *IndexesReader {
	return &IndexesReader{client: client}
}

// SetNoColor set nocolor flag
func (ir *IndexesReader) SetNoColor(nocolor bool) {
	ir.nocolor = nocolor
}

// SetVerbose sets verbose level
func (ir *IndexesReader) SetVerbose(verbose bool) {
	ir.verbose = verbose
}

// SetDBName sets verbose level
func (ir *IndexesReader) SetDBName(dbName string) {
	ir.dbName = dbName
}

// GetIndexes list all indexes of collections of databases
func (ir *IndexesReader) GetIndexes() (bson.M, error) {
	var err error
	var dbNames []string
	indexesMap := bson.M{}
	if ir.dbName != "" {
		indexesMap[ir.dbName], err = ir.GetIndexesFromDB(ir.dbName)
		return indexesMap, err
	}

	if dbNames, err = ListDatabaseNames(ir.client); err != nil {
		return indexesMap, err
	}
	cnt := 0
	for _, name := range dbNames {
		if name == "admin" || name == "config" || name == "local" {
			if ir.verbose == true {
				log.Println("Skip", name)
			}
			continue
		}
		cnt++
		if ir.verbose == true {
			log.Println("checking", name)
		}
		if indexesMap[name], err = ir.GetIndexesFromDB(name); err != nil {
			return indexesMap, err
		}
	}
	if cnt == 0 && ir.verbose == true {
		log.Println("No database is available")
	}
	return indexesMap, err
}

// GetIndexesFromDB list all indexes of collections of a database
func (ir *IndexesReader) GetIndexesFromDB(dbName string) (bson.M, error) {
	var err error
	var cur *mongo.Cursor
	var ctx = context.Background()
	var indexesMap = bson.M{}
	if cur, err = ir.client.Database(dbName).ListCollections(ctx, bson.M{}); err != nil {
		return indexesMap, err
	}
	defer cur.Close(ctx)
	collections := []string{}
	for cur.Next(ctx) {
		var elem = bson.M{}
		if err = cur.Decode(&elem); err != nil {
			if ir.verbose == true {
				log.Println(err)
			}
			continue
		}
		coll := fmt.Sprintf("%v", elem["name"])
		collType := fmt.Sprintf("%v", elem["type"])
		if strings.Index(coll, "system.") == 0 || (elem["type"] != nil && collType != "collection") {
			if ir.verbose == true {
				log.Println("skip", coll)
			}
			continue
		}
		collections = append(collections, coll)
	}

	sort.Strings(collections)
	for _, collection := range collections {
		indexesMap[collection] = ir.GetIndexesFromCollection(ir.client.Database(dbName).Collection(collection))
	}
	return indexesMap, err
}

// GetIndexesFromCollection gets indexes from a collection
func (ir *IndexesReader) GetIndexesFromCollection(collection *mongo.Collection) []IndexStatsDoc {
	var err error
	var ctx = context.Background()
	var pipeline = MongoPipeline(`{"$indexStats": {}}`)
	var list []IndexStatsDoc
	var icur *mongo.Cursor
	var scur *mongo.Cursor
	if ir.verbose {
		log.Println("process GetIndexesFromCollection")
	}

	if scur, err = collection.Aggregate(ctx, pipeline); err != nil {
		if ir.verbose == true {
			log.Println(err)
		}
		return list
	}
	var indexStats = []bson.M{}
	for scur.Next(ctx) {
		var result = bson.M{}
		if err = scur.Decode(&result); err != nil {
			if ir.verbose == true {
				log.Println(err)
			}
			continue
		}
		indexStats = append(indexStats, result)
	}
	scur.Close(ctx)
	indexView := collection.Indexes()
	if icur, err = indexView.List(ctx); err != nil {
		if ir.verbose == true {
			log.Println(err)
		}
		return list
	}
	defer icur.Close(ctx)

	for icur.Next(ctx) {
		var idx = bson.D{}
		if err = icur.Decode(&idx); err != nil {
			if ir.verbose == true {
				log.Println(err)
			}
			continue
		}

		var keys bson.D
		var indexName string
		for _, v := range idx {
			if v.Key == "name" {
				indexName = v.Value.(string)
			} else if v.Key == "key" {
				keys = v.Value.(bson.D)
			}
		}
		var strbuf bytes.Buffer
		fields := []string{}
		for n, value := range keys {
			fields = append(fields, value.Key)
			if n == 0 {
				strbuf.WriteString("{ ")
			}
			strbuf.WriteString(value.Key + ": " + fmt.Sprint(value.Value))
			if n == len(keys)-1 {
				strbuf.WriteString(" }")
			} else {
				strbuf.WriteString(", ")
			}
		}
		o := IndexStatsDoc{Key: strbuf.String(), Fields: fields, Name: indexName}
		// Check shard keys
		var v bson.M
		ns := collection.Database().Name() + "." + collection.Name()
		if ir.verbose {
			log.Println("process", ns)
		}
		if err = ir.client.Database("config").Collection("collections").FindOne(ctx, bson.M{"_id": ns, "key": keys}).Decode(&v); err == nil {
			o.IsShardKey = true
		}
		o.EffectiveKey = strings.Replace(o.Key[2:len(o.Key)-2], ": -1", ": 1", -1)
		o.Usage = []UsageDoc{}
		for _, result := range indexStats {
			if result["name"].(string) == indexName {
				b, _ := bson.Marshal(result)
				var usage UsageDoc
				bson.Unmarshal(b, &usage)
				o.TotalOps += usage.Accesses.Ops
				o.Usage = append(o.Usage, usage)
			}
		}
		list = append(list, o)
	}
	icur.Close(ctx)
	sort.Slice(list, func(i, j int) bool { return (list[i].EffectiveKey < list[j].EffectiveKey) })
	for i, o := range list {
		if o.Key != "{ _id: 1 }" && o.IsShardKey == false {
			list[i].IsDupped = checkIfDupped(o, list)
		}
	}
	return list
}

// check if an index is a dup of others
func checkIfDupped(doc IndexStatsDoc, list []IndexStatsDoc) bool {
	for _, o := range list {
		// check indexes if not marked as dupped, has the same first field, and more or equal number of fields
		if o.IsDupped == false && doc.Fields[0] == o.Fields[0] && doc.Key != o.Key && len(o.Fields) >= len(doc.Fields) {
			nmatched := 0
			for i, fld := range doc.Fields {
				if i == 0 {
					continue
				}
				for j, field := range o.Fields {
					if j > 0 && fld == field {
						nmatched++
						break
					}
				}
			}
			if nmatched == len(doc.Fields)-1 {
				return true
			}
		}
	}
	return false
}

// Print prints indexes
func (ir *IndexesReader) Print(indexesMap bson.M) {
	for _, key := range getSortedKeys(indexesMap) {
		val := indexesMap[key].(bson.M)
		for _, k := range getSortedKeys(val) {
			list := val[k].([]IndexStatsDoc)
			var buffer bytes.Buffer
			ns := key + "." + k
			buffer.WriteString("\n")
			buffer.WriteString(ns)
			buffer.WriteString(":\n")
			for _, o := range list {
				font := codeDefault
				tailCode := codeDefault
				if ir.nocolor {
					font = ""
					tailCode = ""
				}
				if o.Key == "{ _id: 1 }" {
					buffer.WriteString(fmt.Sprintf("%v  %v%v", font, o.Key, tailCode))
				} else if o.IsShardKey == true {
					buffer.WriteString(fmt.Sprintf("%v* %v%v", font, o.Key, tailCode))
				} else if o.IsDupped == true {
					if ir.nocolor == false {
						font = codeRed
					}
					buffer.WriteString(fmt.Sprintf("%vx %v%v", font, o.Key, tailCode))
				} else if o.TotalOps == 0 {
					if ir.nocolor == false {
						font = codeBlue
					}
					buffer.WriteString(fmt.Sprintf("%v? %v%v", font, o.Key, tailCode))
				}

				for _, u := range o.Usage {
					buffer.Write([]byte("\n\thost: " + u.Host + ", ops: " + fmt.Sprintf("%v", u.Accesses.Ops) + ", since: " + fmt.Sprintf("%v", u.Accesses.Since)))
				}
				buffer.WriteString("\n")
			}
			fmt.Println(buffer.String())
		}
	}
}

func getSortedKeys(rmap bson.M) []string {
	var keys []string
	for k := range rmap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
