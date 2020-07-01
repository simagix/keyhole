// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/simagix/gox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DatabaseInfo stores struct
type DatabaseInfo struct {
	redaction bool
	verbose   bool
}

// NewDatabaseInfo returns DatabaseInfo
func NewDatabaseInfo() *DatabaseInfo {
	return &DatabaseInfo{}
}

// SetRedaction sets redaction
func (dbi *DatabaseInfo) SetRedaction(redaction bool) {
	dbi.redaction = redaction
}

// SetVerbose sets verbosity
func (dbi *DatabaseInfo) SetVerbose(verbose bool) {
	dbi.verbose = verbose
}

// GetAllDatabasesInfo gets all db info
func (dbi *DatabaseInfo) GetAllDatabasesInfo(client *mongo.Client) ([]bson.M, error) {
	var err error
	var cur *mongo.Cursor
	var ctx = context.Background()
	var databases = []bson.M{}
	var dbNames []string
	t := time.Now()
	if dbi.verbose {
		log.Println("* GetAllDatabasesInfo")
	}
	if dbNames, err = ListDatabaseNames(client); err != nil {
		if dbi.verbose {
			log.Println(err)
		}
		return databases, err
	}
	// total := len(dbNames)
	for _, dbName := range dbNames {
		// fmt.Fprintf(os.Stderr, "\r%3d%% ", (100*i)/total)
		if dbName == "admin" || dbName == "config" || dbName == "local" {
			if dbi.verbose {
				log.Println("skip database", dbName)
			}
			continue
		}
		if cur, err = client.Database(dbName).ListCollections(ctx, bson.M{}); err != nil {
			if dbi.verbose {
				log.Println(err)
			}
			return databases, err
		}
		defer cur.Close(ctx)
		var collections = []bson.M{}
		ir := NewIndexesReader(client)
		ir.SetVerbose(dbi.verbose)
		collectionNames := []string{}

		for cur.Next(ctx) {
			var elem = bson.M{}
			if err = cur.Decode(&elem); err != nil {
				if dbi.verbose {
					log.Println(err)
				}
				continue
			}
			coll := fmt.Sprintf("%v", elem["name"])
			collType := fmt.Sprintf("%v", elem["type"])
			if strings.Index(coll, "system.") == 0 || (elem["type"] != nil && collType != "collection") {
				if dbi.verbose {
					log.Println("skip", collType, coll)
				}
				continue
			}
			collectionNames = append(collectionNames, coll)
		}

		sort.Strings(collectionNames)
		for _, collectionName := range collectionNames {
			ns := dbName + "." + collectionName
			if dbi.verbose {
				log.Println("GetAllDatabasesInfo", ns)
			}
			collection := client.Database(dbName).Collection(collectionName)

			// firstDoc, FindOne
			var cursor *mongo.Cursor
			var firstDoc bson.M
			opts := options.Find()
			opts.SetLimit(5) // get 5 samples and choose the max_size()
			if cursor, err = collection.Find(ctx, bson.D{{}}, opts); err != nil {
				if dbi.verbose {
					log.Println(err)
				}
				continue
			}
			dsize := 0
			for cursor.Next(ctx) {
				var v bson.M
				cursor.Decode(&v)
				if buf, err := bson.Marshal(v); err != nil {
					if dbi.verbose {
						log.Println(err)
					}
					continue
				} else if len(buf) > dsize {
					firstDoc = v
					dsize = len(buf)
				}
			}
			if firstDoc == nil {
				if dbi.verbose {
					log.Println("no doc available")
				}
				continue
			}
			// firstDoc = emptyBinData(firstDoc)
			if dbi.redaction == true {
				redact := NewRedactor()
				walker := gox.NewMapWalker(redact.callback)
				buf, _ := bson.Marshal(walker.Walk(firstDoc))
				bson.Unmarshal(buf, &firstDoc)
			}

			indexes := ir.GetIndexesFromCollection(collection)

			// stats
			var stats bson.M
			client.Database(dbName).RunCommand(ctx, bson.D{{Key: "collStats", Value: collectionName}}).Decode(&stats)
			chunks := []bson.M{}
			if stats["shards"] != nil {
				for k := range stats["shards"].(primitive.M) {
					m := (stats["shards"].(primitive.M)[k]).(primitive.M)
					delete(m, "$clusterTime")
					delete(m, "$gleStats")
					if chunk, cerr := dbi.collectChunksDistribution(client, k, ns); cerr != nil {
						if dbi.verbose {
							log.Println(cerr)
						}
					} else {
						chunks = append(chunks, chunk)
					}
				}
			}
			collections = append(collections, bson.M{"NS": ns, "collection": collectionName, "chunks": chunks, "document": firstDoc,
				"indexes": indexes, "stats": trimMap(stats)})
		}
		var stats bson.M
		if stats, err = RunCommandOnDB(client, "dbStats", dbName); err != nil {
			if dbi.verbose {
				log.Println(err)
			}
			continue
		}
		databases = append(databases, bson.M{"DB": dbName, "collections": collections, "stats": trimMap(stats)})
	}
	if dbi.verbose {
		log.Println("* GetAllDatabasesInfo took", time.Now().Sub(t))
	}
	return databases, nil
}

func (dbi *DatabaseInfo) collectChunksDistribution(client *mongo.Client, shard string, ns string) (bson.M, error) {
	var count int64
	var ctx = context.Background()
	var cur *mongo.Cursor
	var doc bson.M
	var err error
	var jcount int64
	var ecount int64
	var size int64
	coll := client.Database("config").Collection("collections")
	if err = coll.FindOne(ctx, bson.D{{Key: "_id", Value: ns}, {Key: "dropped", Value: false}}).Decode(&doc); err != nil {
		return doc, err
	}
	t := time.Now()
	coll = client.Database("config").Collection("chunks")
	if dbi.verbose == true {
		log.Println("* collectChunksDistribution on", shard, ns, " ...")
		if cur, err = coll.Find(ctx, bson.M{"ns": ns, "shard": shard}); err != nil {
			return doc, nil
		}
		for cur.Next(ctx) {
			cur.Decode(&doc)
			cmd := bson.D{{Key: "datasize", Value: ns}, {Key: "min", Value: doc["min"]}, {Key: "max", Value: doc["max"]}, {Key: "estimate", Value: true}}
			client.Database("admin").RunCommand(ctx, cmd).Decode(&doc)
			if doc["jumbo"] != nil && doc["jumbo"].(bool) == true {
				jcount++
			}
			if doc["numObjects"] != nil && doc["numObjects"].(float64) == 0 {
				ecount++
			}
			if doc["size"] != nil && doc["size"].(float64) > 0 {
				size += int64(doc["size"].(float64))
			}
			count++
		}
	} else {
		ecount = -1
		size = -1
		if count, err = coll.CountDocuments(ctx, bson.M{"shard": shard, "ns": ns}); err != nil {
			return doc, err
		}
		if jcount, err = coll.CountDocuments(ctx, bson.M{"shard": shard, "ns": ns, "jumbo": true}); err != nil {
			return doc, err
		}
	}
	if dbi.verbose == true {
		dur := time.Now().Sub(t)
		log.Println("* collectChunksDistribution on", shard, ns, " took", dur, "for", count, "chunks, rate:", dur/time.Duration(count))
	}
	return bson.M{"shard": shard, "total": count, "empty": ecount, "jumbo": jcount, "size": size}, err
}
