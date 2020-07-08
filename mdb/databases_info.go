// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/simagix/gox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DatabaseInfo stores struct
type DatabaseInfo struct {
	conns     int
	redaction bool
	verbose   bool
	vv        bool
}

// NewDatabaseInfo returns DatabaseInfo
func NewDatabaseInfo() *DatabaseInfo {
	return &DatabaseInfo{}
}

// SetNumberConnections set # of conns
func (dbi *DatabaseInfo) SetNumberConnections(conns int) {
	dbi.conns = conns
}

// SetRedaction sets redaction
func (dbi *DatabaseInfo) SetRedaction(redaction bool) {
	dbi.redaction = redaction
}

// SetVerbose sets verbosity
func (dbi *DatabaseInfo) SetVerbose(verbose bool) {
	dbi.verbose = verbose
}

// SetVeryVerbose set very verbose
func (dbi *DatabaseInfo) SetVeryVerbose(vv bool) {
	dbi.vv = vv
	if dbi.vv == true {
		dbi.verbose = true
	}
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
		return databases, err
	}
	// total := len(dbNames)
	for _, dbName := range dbNames {
		if dbName == "admin" || dbName == "config" || dbName == "local" {
			if dbi.verbose {
				log.Println("skip database", dbName)
			}
			continue
		}
		if cur, err = client.Database(dbName).ListCollections(ctx, bson.M{}); err != nil {
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
				log.Println(err)
				continue
			}
			dsize := 0
			for cursor.Next(ctx) {
				var v bson.M
				cursor.Decode(&v)
				if buf, err := bson.Marshal(v); err != nil {
					log.Println(err)
					continue
				} else if len(buf) > dsize {
					firstDoc = v
					dsize = len(buf)
				}
			}
			if firstDoc == nil {
				if dbi.verbose {
					log.Println("no sample doc available")
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
				keys := []string{}

				for k := range stats["shards"].(primitive.M) {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					m := (stats["shards"].(primitive.M)[k]).(primitive.M)
					delete(m, "$clusterTime")
					delete(m, "$gleStats")
					if chunk, cerr := dbi.collectChunksDistribution(client, k, ns); cerr != nil {
						// log.Println(cerr)
					} else {
						chunk["objects"] = m["count"]
						chunk["size"] = m["size"]
						chunks = append(chunks, chunk)
					}
				}
			}
			collections = append(collections, bson.M{"NS": ns, "collection": collectionName, "chunks": chunks, "document": firstDoc,
				"indexes": indexes, "stats": trimMap(stats)})
		}
		var stats bson.M
		if stats, err = RunCommandOnDB(client, "dbStats", dbName); err != nil {
			log.Println(err)
			continue
		}
		databases = append(databases, bson.M{"DB": dbName, "collections": collections, "stats": trimMap(stats)})
	}
	if dbi.verbose {
		log.Println("* GetAllDatabasesInfo took", time.Now().Sub(t))
	}
	return databases, nil
}

var batchSize = 16

func (dbi *DatabaseInfo) collectChunksDistribution(client *mongo.Client, shard string, ns string) (bson.M, error) {
	var count int64
	var ctx = context.Background()
	var cur *mongo.Cursor
	var doc bson.D
	var emptyCounts int64
	var err error
	var jumboCounts int64
	var key bson.D
	var mu sync.Mutex
	coll := client.Database("config").Collection("collections")
	if err = coll.FindOne(ctx, bson.D{{Key: "_id", Value: ns}, {Key: "dropped", Value: false}}).Decode(&doc); err != nil {
		return nil, err
	}
	for _, v := range doc {
		if v.Key == "key" {
			key = v.Value.(bson.D)
		}
	}
	t := time.Now()
	coll = client.Database("config").Collection("chunks")
	if dbi.vv == true {
		log.Println("* collectChunksDistribution on", shard, ns, " ...")
		if cur, err = coll.Find(ctx, bson.M{"ns": ns, "shard": shard}); err != nil {
			return nil, nil
		}
		chunks := []bson.M{}
		for cur.Next(ctx) {
			var chunk bson.M
			cur.Decode(&chunk)
			chunks = append(chunks, chunk)
			count++
		}

		var wg = gox.NewWaitGroup(dbi.conns) // runs in parallel
		ptr := 0
		remains := len(chunks)
		for remains > 0 {
			length := batchSize
			if remains < batchSize {
				length = remains
			}
			wg.Add(1)
			go func(chunksBlock []bson.M, key bson.D) {
				defer wg.Done()
				ecount := int64(0)
				jcount := int64(0)
				for _, chunk := range chunksBlock {
					cmd := bson.D{{Key: "datasize", Value: ns}, {Key: "keyPattern", Value: key},
						{Key: "min", Value: chunk["min"]}, {Key: "max", Value: chunk["max"]},
						{Key: "estimate", Value: true}}
					client.Database("admin").RunCommand(ctx, cmd).Decode(&chunk)
					if chunk["jumbo"] != nil && chunk["jumbo"].(bool) == true {
						jcount++
					}
					if chunk["numObjects"] != nil {
						if chunk["numObjects"].(float64) == 0 {
							ecount++
						}
					}
				}
				mu.Lock()
				emptyCounts += ecount
				jumboCounts += jcount
				mu.Unlock()
			}(chunks[ptr:(ptr+length)], key)

			ptr += length
			remains -= length
		}
		wg.Wait()
		dur := time.Now().Sub(t)
		log.Println("* collectChunksDistribution used", dbi.conns, "threads on", shard, ns, "took", dur, "for", count, "chunks, rate:", dur/time.Duration(count))
	} else {
		emptyCounts = -1
		if count, err = coll.CountDocuments(ctx, bson.M{"shard": shard, "ns": ns}); err != nil {
			return nil, err
		}
		if jumboCounts, err = coll.CountDocuments(ctx, bson.M{"shard": shard, "ns": ns, "jumbo": true}); err != nil {
			return nil, err
		}
	}
	info := bson.M{"shard": shard, "total": count, "empty": emptyCounts, "jumbo": jumboCounts}
	return info, err
}
