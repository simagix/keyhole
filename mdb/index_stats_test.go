// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"log"
	"math/rand"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var dbName = "keyhole"

func TestGetIndexesFromDB(t *testing.T) {
	var client = getMongoClient()
	defer client.Disconnect(context.Background())
	c := client.Database(dbName).Collection(ExamplesCollection)
	seedNumbers(c)

	// get index from keyhole database
	ir := NewIndexStats("utest-xxxxxx")
	str, _ := ir.GetIndexesFromDB(client, dbName)
	t.Log(str)
}

func TestGetIndexes(t *testing.T) {
	var client = getMongoClient()
	defer client.Disconnect(context.Background())
	c := client.Database(dbName).Collection(ExamplesCollection)
	seedNumbers(c)

	// get all indexes
	ir := NewIndexStats("utest-xxxxxx")
	str, _ := ir.GetIndexes(client)
	t.Log(str)
}

func seedNumbers(c *mongo.Collection) {
	var err error
	var ctx = context.Background()
	c.Drop(ctx)

	var docs []interface{}
	for n := 0; n < 1000; n++ {
		docs = append(docs, bson.M{"a": rand.Intn(100), "b": rand.Intn(100), "c": rand.Intn(100)})
	}
	if _, err = c.InsertMany(ctx, docs); err != nil {
		log.Fatal(err)
	}
	// create index example
	indexView := c.Indexes()
	idx := mongo.IndexModel{
		Keys: bson.D{{Key: "a", Value: 1}},
	}
	indexView.CreateOne(ctx, idx)
	idx = mongo.IndexModel{
		Keys: bson.D{{Key: "a", Value: 1}, {Key: "b", Value: 1}},
	}
	indexView.CreateOne(ctx, idx)
	idx = mongo.IndexModel{
		Keys: bson.D{{Key: "a", Value: 1}, {Key: "b", Value: -1}},
	}
	indexView.CreateOne(ctx, idx)
	idx = mongo.IndexModel{
		Keys: bson.D{{Key: "a", Value: 1}, {Key: "b", Value: 1}, {Key: "c", Value: 1}},
	}
	indexView.CreateOne(ctx, idx)
}
