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

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
)

// UsageDoc -
type UsageDoc struct {
	Hostname string
	Ops      int       `json:"ops" bson:"ops"`
	Since    time.Time `json:"since" bson:"since"`
}

// IndexStatsDoc -
type IndexStatsDoc struct {
	Key          string
	EffectiveKey string
	Usage        []UsageDoc
}

// GetIndexes list all indexes of collections of databases
func GetIndexes(client *mongo.Client, dbName string) string {
	if dbName != "" {
		return GetIndexesFromDB(client, dbName)
	}

	var buffer bytes.Buffer
	dbNames, _ := ListDatabaseNames(client)
	for _, name := range dbNames {
		if name == "admin" || name == "config" || name == "local" {
			continue
		}
		buffer.WriteString(GetIndexesFromDB(client, name))
	}
	return buffer.String()
}

// GetIndexesFromDB list all indexes of collections of a database
func GetIndexesFromDB(client *mongo.Client, dbName string) string {
	var err error
	var cur *mongo.Cursor
	var icur *mongo.Cursor
	var scur *mongo.Cursor
	var buffer bytes.Buffer
	var ctx = context.Background()
	// var pipeline = mongo.Pipeline{{{Key: "$indexStats", Value: bson.M{}}}}
	var pipeline = MongoPipeline(`{"$indexStats": {}}`)
	if cur, err = client.Database(dbName).ListCollections(ctx, bson.M{}); err != nil {
		return buffer.String()
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var elem = bson.M{}
		if err = cur.Decode(&elem); err != nil {
			continue
		}
		coll := fmt.Sprintf("%v", elem["name"])
		collType := fmt.Sprintf("%v", elem["type"])
		if strings.Index(coll, "system.") == 0 || collType != "collection" {
			continue
		}

		buffer.WriteString("\n")
		buffer.WriteString(dbName)
		buffer.WriteString(".")
		buffer.WriteString(coll)
		buffer.WriteString(":\n")

		if scur, err = client.Database(dbName).Collection(coll).Aggregate(ctx, pipeline); err != nil {
			log.Fatal(err)
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
		indexView := client.Database(dbName).Collection(coll).Indexes()
		if icur, err = indexView.List(ctx); err != nil {
			continue
		}
		defer icur.Close(ctx)
		var list []IndexStatsDoc

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
			o := IndexStatsDoc{Key: strbuf.String()}
			o.EffectiveKey = strings.Replace(o.Key[:len(o.Key)-2], ": -1", ": 1", -1)
			o.Usage = []UsageDoc{}
			for _, result := range indexStats {
				if result["name"].(string) == indexName {
					doc := result["accesses"].(bson.M)
					host := result["host"].(string)
					b, _ := bson.Marshal(doc)
					var accesses UsageDoc
					bson.Unmarshal(b, &accesses)
					accesses.Hostname = host
					o.Usage = append(o.Usage, accesses)
				}
			}
			list = append(list, o)
		}
		icur.Close(ctx)
		sort.Slice(list, func(i, j int) bool { return (list[i].EffectiveKey <= list[j].EffectiveKey) })
		for i, o := range list {
			font := "\x1b[0m  "
			if o.Key != "{ _id: 1 }" {
				if i < len(list)-1 && strings.Index(list[i+1].EffectiveKey, o.EffectiveKey) == 0 {
					font = "\x1b[31;1mx " // red
				} else {
					sum := 0
					for _, u := range o.Usage {
						sum += u.Ops
					}
					if sum == 0 {
						font = "\x1b[34;1m? " // blue
					}
				}
			}
			buffer.WriteString(font + o.Key + "\x1b[0m")
			for _, u := range o.Usage {
				buffer.Write([]byte("\n\thost: " + u.Hostname + ", ops: " + fmt.Sprintf("%v", u.Ops) + ", since: " + fmt.Sprintf("%v", u.Since)))
			}
			buffer.WriteString("\n")
		}
	}
	return buffer.String()
}
