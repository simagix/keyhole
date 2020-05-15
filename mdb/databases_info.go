// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DatabaseInfo stores struct
type DatabaseInfo struct {
	verbose bool
}

// NewDatabaseInfo returns DatabaseInfo
func NewDatabaseInfo() *DatabaseInfo {
	return &DatabaseInfo{}
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
	if dbi.verbose {
		log.Println("process GetAllDatabasesInfo")
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
				log.Println("process", ns)
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
			indexes := ir.GetIndexesFromCollection(collection)

			// stats
			var stats bson.M
			client.Database(dbName).RunCommand(ctx, bson.D{{Key: "collStats", Value: collectionName}}).Decode(&stats)
			if stats["shards"] != nil {
				for k := range stats["shards"].(primitive.M) {
					m := (stats["shards"].(primitive.M)[k]).(primitive.M)
					delete(m, "$clusterTime")
					delete(m, "$gleStats")
				}
			}
			collections = append(collections, bson.M{"NS": ns, "collection": collectionName, "document": firstDoc,
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
	return databases, nil
}
