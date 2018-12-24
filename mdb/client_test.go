// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
)

func TestMongoClient(t *testing.T) {
	var err error
	var client *mongo.Client
	// var info ServerInfo

	uri := "mongodb://localhost/keyhole"
	if os.Getenv("DATABASE_URL") != "" {
		uri = os.Getenv("DATABASE_URL")
	}

	if client, err = mongo.NewClient(uri); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err = client.Connect(ctx); err != nil {
		panic(err)
	}
	var cur mongo.Cursor
	if cur, err = client.Database(dbName).ListCollections(ctx, bson.M{}); err != nil {
		t.Fatal(err)
	}
	pipeline := mongo.Pipeline{{{Key: "$indexStats", Value: bson.M{}}}}
	var scur mongo.Cursor
	if scur, err = client.Database("keyhole").Collection("examples").Aggregate(ctx, pipeline); err != nil {
		log.Fatal(err)
	}

	for cur.Next(ctx) {
		var coll = bson.M{}
		if err = cur.Decode(&coll); err != nil {
			fmt.Println(err)
			continue
		}
		// fmt.Println(coll)
		for scur.Next(ctx) {
			var idx = bson.M{}
			if err = scur.Decode(&idx); err != nil {
				fmt.Println(err)
				continue
			}
			// fmt.Println(idx)
		}
	}
}
