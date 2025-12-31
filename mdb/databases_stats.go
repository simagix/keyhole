// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"fmt"
	"runtime"
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

const (
	maxNumShards       = 3
	minWorkers         = 4
	sampleDocSizeLimit = (32 * 1024) // 32KB
)

// getNumWorkers returns the number of parallel workers (max of NumCPU and minWorkers)
func getNumWorkers() int {
	n := runtime.NumCPU()
	if n < minWorkers {
		return minWorkers
	}
	return n
}

// DatabaseStats stores struct
type DatabaseStats struct {
	Logger *gox.Logger

	fastMode  bool
	numShards int
	redaction bool
	threads   int
	verbose   bool
	version   string
}

// ListDatabases stores listDatabases
type ListDatabases struct {
	Databases   []Database `bson:"databases"`
	TotalSize   int64      `bson:"totalSize"`
	TotalSizeMB int64      `bson:"totalSizeMb"`
}

// Database stores database stats
type Database struct {
	Name        string       `bson:"name"`
	Collections []Collection `bson:"collections,truncate"`
	SizeOnDisk  int64        `bson:"sizeOnDisk,truncate"`
	Empty       bool         `bson:"empty,truncate"`
	Shards      bson.M       `bson:"shards,truncate"`
	Stats       struct {
		Raw         bson.M `bson:"raw,truncate"`
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

// Collection stores struct
type Collection struct {
	Chunks   []Chunk `bson:"chunks,truncate"`
	Document bson.M  `bson:"document,truncate"`
	Indexes  []Index `bson:"indexes,truncate"`
	Name     string  `bson:"name,truncate"`
	NS       string  `bson:"namespace,truncate"`
	Stats    struct {
		Count          int64   `bson:"count,truncate"`
		IndexDetails   bson.M  `bson:"indexDetails,truncate"`
		AvgObjSize     float64 `bson:"avgObjSize,truncate"`
		WiredTiger     bson.M  `bson:"wiredTiger,truncate"`
		IndexSizes     bson.M  `bson:"indexSizes,truncate"`
		Capped         bool    `bson:"capped,truncate"`
		TotalIndexSize int64   `bson:"totalIndexSize,truncate"`
		MaxSize        int64   `bson:"maxSize,truncate"`
		Nindexes       int64   `bson:"nindexes,truncate"`
		Shards         bson.M  `bson:"shards,truncate"`
		NS             string  `bson:"namespace,truncate"`
		Nchunks        int64   `bson:"nchunks,truncate"`
		Sharded        bool    `bson:"sharded,truncate"`
		Size           int64   `bson:"size,truncate"`
		StorageSize    int64   `bson:"storageSize,truncate"`
	} `bson:"stats"`
}

// Chunk stores chunk stats
type Chunk struct {
	Empty   int64  `bson:"empty"`
	Jumbo   int64  `bson:"jumbo"`
	Objects int64  `bson:"objects"`
	Shard   string `bson:"shard"`
	Size    int64  `bson:"size"`
	Total   int64  `bson:"total"`
}

// NewDatabaseStats returns DatabaseStats
func NewDatabaseStats(version string) *DatabaseStats {
	return &DatabaseStats{Logger: gox.GetLogger(version), threads: 16, version: version}
}

// SetFastMode sets fastMode mode
func (p *DatabaseStats) SetFastMode(fastMode bool) {
	p.fastMode = fastMode
}

// SetNumberShards set # of threads
func (p *DatabaseStats) SetNumberShards(n int) {
	p.numShards = n
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
	if verbose && p.Logger != nil {
		p.Logger.SetLoggerLevel(gox.Debug)
	}
}

// GetAllDatabasesStats gets all db info
func (p *DatabaseStats) GetAllDatabasesStats(client *mongo.Client, dbNames []string) ([]Database, error) {
	var err error
	var ctx = context.Background()
	var databases []Database
	t := time.Now()
	p.Logger.Debug("GetAllDatabasesStats")

	dbList, err := p.getDatabaseList(client, dbNames)
	if err != nil {
		return nil, err
	}

	for _, db := range dbList {
		collections, err := p.getCollectionsForDatabase(client, db.Name)
		if err != nil {
			p.Logger.Error(err.Error())
			continue
		}
		if err = client.Database(db.Name).RunCommand(ctx, bson.D{{Key: "dbStats", Value: 1}}).Decode(&db.Stats); err != nil {
			p.Logger.Error(err.Error())
			continue
		}
		db.Collections = collections
		databases = append(databases, db)
	}
	p.Logger.Debugf("GetAllDatabasesStats took %v", time.Since(t))
	return databases, nil
}

// GetDatabasesSummaries returns database metadata without collections (lightweight)
func (p *DatabaseStats) GetDatabasesSummaries(client *mongo.Client, dbNames []string) ([]Database, error) {
	var err error
	var ctx = context.Background()
	var databases []Database
	p.Logger.Debug("GetDatabasesSummaries")

	dbList, err := p.getDatabaseList(client, dbNames)
	if err != nil {
		return nil, err
	}

	for _, db := range dbList {
		if err = client.Database(db.Name).RunCommand(ctx, bson.D{{Key: "dbStats", Value: 1}}).Decode(&db.Stats); err != nil {
			p.Logger.Error(err.Error())
			continue
		}
		databases = append(databases, db)
	}
	return databases, nil
}

// CollectionWriter is a callback for streaming collections
type CollectionWriter func(coll Collection) error

// StreamCollections collects and streams each collection via callback
func (p *DatabaseStats) StreamCollections(client *mongo.Client, dbNames []string, writer CollectionWriter) error {
	p.Logger.Debug("StreamCollections")

	dbList, err := p.getDatabaseList(client, dbNames)
	if err != nil {
		return err
	}

	for _, db := range dbList {
		collections, err := p.getCollectionsForDatabase(client, db.Name)
		if err != nil {
			p.Logger.Error(err.Error())
			continue
		}
		// Stream each collection immediately
		for _, coll := range collections {
			if err := writer(coll); err != nil {
				return err
			}
		}
	}
	return nil
}

// getDatabaseList returns filtered and sorted database list
func (p *DatabaseStats) getDatabaseList(client *mongo.Client, dbNames []string) ([]Database, error) {
	var ctx = context.Background()
	var listdb ListDatabases

	if err := client.Database("admin").RunCommand(ctx, bson.D{{Key: "listDatabases", Value: 1}}).Decode(&listdb); err != nil {
		return nil, err
	}

	if len(dbNames) > 0 {
		dict := map[string]bool{}
		for _, name := range dbNames {
			dict[name] = true
		}
		list := []Database{}
		for _, db := range listdb.Databases {
			if dict[db.Name] {
				list = append(list, db)
			}
		}
		listdb.Databases = list
	}

	sort.Slice(listdb.Databases, func(i, j int) bool {
		return listdb.Databases[i].Name < listdb.Databases[j].Name
	})

	// Filter out system databases
	var result []Database
	for _, db := range listdb.Databases {
		if db.Name == "admin" || db.Name == "config" || db.Name == "local" {
			continue
		}
		result = append(result, db)
	}
	return result, nil
}

// getCollectionsForDatabase collects all collections for a single database
func (p *DatabaseStats) getCollectionsForDatabase(client *mongo.Client, dbName string) ([]Collection, error) {
	var err error
	var cur *mongo.Cursor
	var ctx = context.Background()
	var collections []Collection

	if cur, err = client.Database(dbName).ListCollections(ctx, bson.D{{}}); err != nil {
		return nil, err
	}

	ir := NewIndexStats(p.version)
	ir.SetVerbose(p.verbose)
	ir.SetFastMode(p.fastMode)
	collectionNames := []string{}

	for cur.Next(ctx) {
		var elem = bson.M{}
		if err = cur.Decode(&elem); err != nil {
			continue
		}
		coll := fmt.Sprintf("%v", elem["name"])
		collType := fmt.Sprintf("%v", elem["type"])
		if collType != "timeseries" && collType != "collection" {
			p.Logger.Debugf(`skip %v %v`, collType, coll)
			continue
		}
		collectionNames = append(collectionNames, coll)
	}
	cur.Close(ctx)

	sort.Strings(collectionNames)
	var wg = gox.NewWaitGroup(getNumWorkers()) // parallel workers based on CPU count
	var mu sync.Mutex
	for _, collectionName := range collectionNames {
		wg.Add(1)
		go func(client *mongo.Client, collectionName string) {
			defer wg.Done()
			ns := dbName + "." + collectionName
			p.Logger.Debugf(`collecting from %v`, ns)
			collection := client.Database(dbName).Collection(collectionName)

			var cursor *mongo.Cursor
			var sampleDoc bson.M
			opts := options.Find()
			opts.SetLimit(3) // get up to 3 docs, prefer the 3rd one
			opts.SetHint(bson.D{{Key: "$natural", Value: 1}})

			if !strings.HasPrefix(collectionName, "system.") {
				if cursor, err = collection.Find(ctx, bson.D{{}}, opts); err != nil {
					p.Logger.Error(err.Error())
					return
				}

				// Collect up to 3 docs, use the 3rd if available, otherwise last available
				var docs []bson.M
				for cursor.Next(ctx) && len(docs) < 3 {
					var v bson.M
					cursor.Decode(&v)
					docs = append(docs, v)
				}
				cursor.Close(ctx)

				if len(docs) > 0 {
					// Prefer 3rd doc (index 2), fallback to last available
					idx := len(docs) - 1
					if idx > 2 {
						idx = 2
					}
					candidate := docs[idx]
					if buf, err := bson.Marshal(candidate); err != nil {
						p.Logger.Error(err.Error())
					} else if len(buf) < sampleDocSizeLimit {
						sampleDoc = candidate
					} else {
						sampleDoc = bson.M{"warning": "sample doc collecting skipped because doc size exceeds 32KB"}
					}
				}

				if sampleDoc == nil {
					p.Logger.Debug("no sample doc available")
				}
			} else {
				p.Logger.Debug("skip ", collectionName)
			}
			if p.redaction {
				redact := NewRedactor()
				walker := gox.NewMapWalker(redact.callback)
				buf, _ := bson.Marshal(walker.Walk(sampleDoc))
				bson.Unmarshal(buf, &sampleDoc)
			}
			indexes, err := ir.GetIndexesFromCollection(client, collection)
			if err != nil {
				p.Logger.Error(err)
			}
			collstats := Collection{NS: ns, Name: collectionName, Document: sampleDoc, Indexes: indexes}
			chunks := []Chunk{}
			if !p.fastMode {
				// Decode directly into struct (avoid marshal/unmarshal roundtrip)
				client.Database(dbName).RunCommand(ctx, bson.D{{Key: "collStats", Value: collectionName}}).Decode(&collstats.Stats)
				if collstats.Stats.Shards != nil {
					shardNames := []string{}
					for shard := range collstats.Stats.Shards {
						shardNames = append(shardNames, shard)
					}
					sort.Strings(shardNames)
					for _, k := range shardNames {
						m := (collstats.Stats.Shards[k]).(primitive.M)
						delete(m, "$clusterTime")
						delete(m, "$gleStats")
						if chunk, cerr := p.collectChunksDistribution(client, k, ns); cerr != nil {
							p.Logger.Errorf(`ns %v error %v`, ns, cerr)
						} else {
							chunk.Objects = toInt64(m["count"])
							chunk.Size = toInt64(m["size"])
							chunks = append(chunks, chunk)
						}
					}
				}
			}
			collstats.Chunks = chunks
			mu.Lock()
			collections = append(collections, collstats)
			mu.Unlock()
		}(client, collectionName)
	}
	wg.Wait()
	sort.Slice(collections, func(i, j int) bool {
		return collections[i].Name < collections[j].Name
	})
	return collections, nil
}

func (p *DatabaseStats) collectChunksDistribution(client *mongo.Client, shard string, ns string) (Chunk, error) {
	var batchSize = 5
	var count int64
	var ctx = context.Background()
	var cur *mongo.Cursor
	var doc bson.D
	var emptyCounts int64
	var err error
	var jumboCounts int64
	var key bson.D
	var chunk Chunk
	var mu sync.Mutex
	coll := client.Database("config").Collection("collections")
	if err = coll.FindOne(ctx, bson.D{{Key: "_id", Value: ns}, {Key: "dropped", Value: bson.M{"$ne": true}}}).Decode(&doc); err != nil {
		return chunk, nil
	}
	for _, v := range doc {
		if v.Key == "key" {
			key = v.Value.(bson.D)
		}
	}
	t := time.Now()
	coll = client.Database("config").Collection("chunks")
	if p.verbose {
		p.Logger.Info(fmt.Sprintf(`collectChunksDistribution on %v %v ...`, shard, ns))
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
					if chunk["jumbo"] != nil && chunk["jumbo"].(bool) {
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
		dur := time.Since(t)
		msg := fmt.Sprintf("collectChunksDistribution used %v threads on %v %v took %v for %v chunks, rate: %v",
			p.threads, shard, ns, dur, count, dur/time.Duration(count))
		p.Logger.Info(msg)
	} else {
		emptyCounts = -1
		if count, err = coll.CountDocuments(ctx, bson.M{"shard": shard, "ns": ns}); err != nil {
			return chunk, err
		}
		if jumboCounts, err = coll.CountDocuments(ctx, bson.M{"shard": shard, "ns": ns, "jumbo": true}); err != nil {
			return chunk, err
		}
	}
	chunk = Chunk{Shard: shard, Total: count, Empty: emptyCounts, Jumbo: jumboCounts}
	return chunk, err
}
