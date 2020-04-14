// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetAllDatabasesInfo gets all db info
func GetAllDatabasesInfo(client *mongo.Client) ([]bson.M, error) {
	var err error
	var cur *mongo.Cursor
	var ctx = context.Background()
	var databases = []bson.M{}
	var dbNames []string
	if dbNames, err = ListDatabaseNames(client); err != nil {
		return databases, err
	}
	// total := len(dbNames)
	for _, dbName := range dbNames {
		// fmt.Fprintf(os.Stderr, "\r%3d%% ", (100*i)/total)
		if dbName == "admin" || dbName == "config" || dbName == "local" {
			continue
		}
		if cur, err = client.Database(dbName).ListCollections(ctx, bson.M{}); err != nil {
			return databases, err
		}
		defer cur.Close(ctx)
		var collections = []bson.M{}
		ir := NewIndexesReader(client)
		collectionNames := []string{}

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
			collectionNames = append(collectionNames, coll)
		}

		sort.Strings(collectionNames)
		for _, collectionName := range collectionNames {
			ns := dbName + "." + collectionName
			collection := client.Database(dbName).Collection(collectionName)

			// firstDoc, FindOne
			var firstDoc bson.M
			if err = collection.FindOne(ctx, bson.D{{}}).Decode(&firstDoc); err != nil {
				continue
			}
			firstDoc = emptyBinData(firstDoc)
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
			continue
		}
		databases = append(databases, bson.M{"DB": dbName, "collections": collections, "stats": trimMap(stats)})
	}
	return databases, nil
}
