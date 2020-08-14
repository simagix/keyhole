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

// DatabaseStats stores struct
type DatabaseStats struct {
	threads   int
	logs      []string
	redaction bool
	verbose   bool
	vv        bool
}

// ListDatabases stores listDatabases
type ListDatabases struct {
	Databases   []Database `bson:"databases"`
	TotalSize   int64      `bson:"totalSize"`
	TotalSizeMB int64      `bson:"totalSizeMb"`
}

// Database stores database stats
type Database struct {
	Name        string            `bson:"name"`
	Collections []CollectionStats `bson:"collections"`
	DB          string            `bson:"DB"` // decoy
	SizeOnDisk  int64             `bson:"sizeOnDisk,truncate"`
	Empty       bool              `bson:"empty"`
	Shards      bson.M            `bson:"shards"`
	Stats       struct {
		Raw         bson.M `bson:"raw"`
		Objects     int64  `bson:"objects,truncate"`
		IndexSize   int64  `bson:"indexSize,truncate"`
		FileSize    int64  `bson:"fileSize,truncate"`
		Indexes     int64  `bson:"indexes,truncate"`
		ScaleFactor int64  `bson:"scaleFactor,truncate"`
		AvgObjSize  int64  `bson:"avgObjSize,truncate"`
		DataSize    int64  `bson:"dataSize,truncate"`
		StorageSize int64  `bson:"storageSize,truncate"`
		NumExtents  int64  `bson:"numExtents,truncate"`
	}
}

// NewDatabaseStats returns DatabaseStats
func NewDatabaseStats() *DatabaseStats {
	return &DatabaseStats{logs: []string{}, threads: 16, vv: true}
}

// GetLogs returns logs
func (p *DatabaseStats) GetLogs() []string {
	return p.logs
}

// SetNumberThreads set # of threads
func (p *DatabaseStats) SetNumberThreads(threads int) {
	p.threads = threads
}

// SetRedaction sets redaction
func (p *DatabaseStats) SetRedaction(redaction bool) {
	p.redaction = redaction
}

// SetVerbose sets verbosity
func (p *DatabaseStats) SetVerbose(verbose bool) {
	p.verbose = verbose
}

// SetVeryVerbose set very verbose
func (p *DatabaseStats) SetVeryVerbose(vv bool) {
	p.vv = vv
	if p.vv == true {
		p.verbose = true
	}
}

// GetAllDatabasesStats gets all db info
func (p *DatabaseStats) GetAllDatabasesStats(client *mongo.Client) ([]Database, error) {
	var err error
	var cur *mongo.Cursor
	var ctx = context.Background()
	var listdb ListDatabases
	var databases []Database
	t := time.Now()
	log.Println("GetAllDatabasesStats")
	if err = client.Database("admin").RunCommand(ctx, bson.D{{Key: "listDatabases", Value: 1}}).Decode(&listdb); err != nil {
		return listdb.Databases, nil
	}
	for _, db := range listdb.Databases {
		if db.Name == "admin" || db.Name == "config" || db.Name == "local" {
			log.Println("skip database " + db.Name)
			continue
		}
		if cur, err = client.Database(db.Name).ListCollections(ctx, bson.D{{}}); err != nil {
			return listdb.Databases, err
		}
		defer cur.Close(ctx)
		var collections = []CollectionStats{}
		ir := NewIndexes(client)
		ir.SetVerbose(p.verbose)
		collectionNames := []string{}

		for cur.Next(ctx) {
			var elem = bson.M{}
			if err = cur.Decode(&elem); err != nil {
				continue
			}
			coll := fmt.Sprintf("%v", elem["name"])
			collType := fmt.Sprintf("%v", elem["type"])
			if strings.Index(coll, "system.") == 0 || (elem["type"] != nil && collType != "collection") {
				if p.verbose {
					log.Println(fmt.Sprintf(`skip %v %v`, collType, coll))
				}
				continue
			}
			collectionNames = append(collectionNames, coll)
		}

		sort.Strings(collectionNames)
		var wg = gox.NewWaitGroup(4) // runs in parallel
		var mu sync.Mutex
		for _, collectionName := range collectionNames {
			wg.Add(1)
			go func(client *mongo.Client, collectionName string) {
				defer wg.Done()
				ns := db.Name + "." + collectionName
				msg := fmt.Sprintf(`collecting from %v`, ns)
				p.logs = append(p.logs, msg)
				log.Println(msg)
				collection := client.Database(db.Name).Collection(collectionName)

				var cursor *mongo.Cursor
				var sampleDoc bson.M
				opts := options.Find()
				opts.SetLimit(5) // get 5 samples and choose the max_size()
				if cursor, err = collection.Find(ctx, bson.D{{}}, opts); err != nil {
					log.Println(err.Error())
					return
				}
				dsize := 0
				for cursor.Next(ctx) {
					var v bson.M
					cursor.Decode(&v)
					if buf, err := bson.Marshal(v); err != nil {
						log.Println(err.Error())
						continue
					} else if len(buf) > dsize {
						sampleDoc = v
						dsize = len(buf)
					}
				}
				if sampleDoc == nil {
					if p.verbose {
						log.Println("no sample doc available")
					}
				}
				if p.redaction == true {
					redact := NewRedactor()
					walker := gox.NewMapWalker(redact.callback)
					buf, _ := bson.Marshal(walker.Walk(sampleDoc))
					bson.Unmarshal(buf, &sampleDoc)
				}
				indexes, err := ir.GetIndexesFromCollection(collection)
				if err != nil {
					return
				}

				// stats
				var stats bson.M
				client.Database(db.Name).RunCommand(ctx, bson.D{{Key: "collStats", Value: collectionName}}).Decode(&stats)
				chunks := []ChunkStats{}
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
						if chunk, cerr := p.collectChunksDistribution(client, k, ns); cerr != nil {
							// log.Println(cerr)
						} else {
							chunk.Objects = toInt64(m["count"])
							chunk.Size = toInt64(m["size"])
							chunks = append(chunks, chunk)
						}
					}
				}
				mu.Lock()
				collstats := CollectionStats{NS: ns, Collection: collectionName, Chunks: chunks, Document: sampleDoc,
					Indexes: indexes}
				data, _ := bson.Marshal(stats)
				bson.Unmarshal(data, &collstats.Stats)
				collections = append(collections, collstats)
				mu.Unlock()
			}(client, collectionName)
		}
		wg.Wait()
		sort.Slice(collections, func(i, j int) bool {
			return collections[i].Collection < collections[j].Collection
		})
		if err = client.Database(db.Name).RunCommand(ctx, bson.D{{Key: "dbStats", Value: 1}}).Decode(&db.Stats); err != nil {
			log.Println(err.Error())
			continue
		}
		db.Collections = collections
		databases = append(databases, db)
	}
	msg := fmt.Sprintf("GetAllDatabasesInfo took %v", time.Now().Sub(t))
	p.logs = append(p.logs, msg)
	log.Println(msg)
	return databases, nil
}

func (p *DatabaseStats) collectChunksDistribution(client *mongo.Client, shard string, ns string) (ChunkStats, error) {
	var batchSize = 5
	var count int64
	var ctx = context.Background()
	var cur *mongo.Cursor
	var doc bson.D
	var emptyCounts int64
	var err error
	var jumboCounts int64
	var key bson.D
	var chunk ChunkStats
	var mu sync.Mutex
	coll := client.Database("config").Collection("collections")
	if err = coll.FindOne(ctx, bson.D{{Key: "_id", Value: ns}, {Key: "dropped", Value: false}}).Decode(&doc); err != nil {
		return chunk, err
	}
	for _, v := range doc {
		if v.Key == "key" {
			key = v.Value.(bson.D)
		}
	}
	t := time.Now()
	coll = client.Database("config").Collection("chunks")
	if p.vv == true {
		log.Println(fmt.Sprintf(`collectChunksDistribution on %v %v ...`, shard, ns))
		if cur, err = coll.Find(ctx, bson.M{"ns": ns, "shard": shard}); err != nil {
			return chunk, nil
		}
		chunks := []bson.M{}
		for cur.Next(ctx) {
			var chunk bson.M
			cur.Decode(&chunk)
			chunks = append(chunks, chunk)
			count++
		}

		var wg = gox.NewWaitGroup(p.threads) // runs in parallel
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
						if toInt64(chunk["numObjects"]) == 0 {
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
		msg := fmt.Sprintf("collectChunksDistribution used %v threads on %v %v took %v for %v chunks, rate: %v",
			p.threads, shard, ns, dur, count, dur/time.Duration(count))
		p.logs = append(p.logs, msg)
		log.Println(msg)
	} else {
		emptyCounts = -1
		if count, err = coll.CountDocuments(ctx, bson.M{"shard": shard, "ns": ns}); err != nil {
			return chunk, err
		}
		if jumboCounts, err = coll.CountDocuments(ctx, bson.M{"shard": shard, "ns": ns, "jumbo": true}); err != nil {
			return chunk, err
		}
	}
	chunk = ChunkStats{Shard: shard, Total: count, Empty: emptyCounts, Jumbo: jumboCounts}
	return chunk, err
}
