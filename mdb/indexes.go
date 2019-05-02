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
	indexesMap := bson.M{}
	if ir.dbName != "" {
		indexesMap[ir.dbName], err = ir.GetIndexesFromDB(ir.dbName)
		return indexesMap, err
	}

	dbNames, _ := ListDatabaseNames(ir.client)
	for _, name := range dbNames {
		if name == "admin" || name == "config" || name == "local" {
			continue
		}
		if indexesMap[name], err = ir.GetIndexesFromDB(name); err != nil {
			return indexesMap, err
		}
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
			continue
		}
		coll := fmt.Sprintf("%v", elem["name"])
		collType := fmt.Sprintf("%v", elem["type"])
		if strings.Index(coll, "system.") == 0 || (elem["type"] != nil && collType != "collection") {
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

	if scur, err = collection.Aggregate(ctx, pipeline); err != nil {
		log.Fatal(err)
		return list
	}
	var indexStats = []bson.M{}
	for scur.Next(ctx) {
		var result = bson.M{}
		if err = scur.Decode(&result); err != nil {
			continue
		}
		indexStats = append(indexStats, result)
	}
	scur.Close(ctx)
	indexView := collection.Indexes()
	if icur, err = indexView.List(ctx); err != nil {
		return list
	}
	defer icur.Close(ctx)

	for icur.Next(ctx) {
		var idx = bson.D{}
		if err = icur.Decode(&idx); err != nil {
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
		for n, value := range keys {
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
		o := IndexStatsDoc{Key: strbuf.String(), Name: indexName}
		// TODO
		var v bson.M
		ns := collection.Database().Name() + "." + collection.Name()
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
	sort.Slice(list, func(i, j int) bool { return (list[i].EffectiveKey <= list[j].EffectiveKey) })
	for i, o := range list {
		if o.Key != "{ _id: 1 }" && o.IsShardKey == false {
			if i < len(list)-1 && strings.Index(list[i+1].EffectiveKey, o.EffectiveKey) == 0 {
				list[i].IsDupped = true
			}
		}
	}
	return list
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
				font := "\x1b[0m  "
				if o.Key == "{ _id: 1 }" {
				} else if o.IsShardKey == true {
					font = "\x1b[0m* "
				} else if o.IsDupped == true {
					font = "\x1b[31;1mx " // red
				} else if o.TotalOps == 0 {
					font = "\x1b[34;1m? " // blue
				}

				buffer.WriteString(font + o.Key + "\x1b[0m")
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
