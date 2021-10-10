// Copyright 2021 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"context"
	"errors"
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	UnitTestURL = "mongodb://localhost/?replicaSet=replset"
)

func GetMongoClient() *mongo.Client {
	var err error
	var client *mongo.Client

	if os.Getenv("DATABASE_URL") != "" {
		UnitTestURL = os.Getenv("DATABASE_URL")
	}
	if client, err = mongo.Connect(context.Background(), options.Client().ApplyURI(UnitTestURL)); err != nil {
		panic(err)
	}

	return client
}

func TestSeed(t *testing.T) {
	var err error
	var client = GetMongoClient()
	defer client.Disconnect(context.Background())
	dbName := "keyhole"
	total := 100
	f := NewSeed()
	f.SetDatabase(dbName)
	f.SetTotal(total)
	f.SetIsDrop(true)
	f.SetShowProgress(false)

	// calling seed()
	f.SeedData(client)
	db := client.Database(dbName)
	coll := db.Collection("vehicles")
	var count int64
	if count, err = coll.CountDocuments(context.Background(), bson.M{}); err != nil {
		t.Fatal(err)
	}

	if int64(total) != count {
		t.Fatal(errors.New("vehicles count doesn't match"))
	}
}

func TestSeedFromTemplate(t *testing.T) {
	var err error
	var client *mongo.Client
	ctx := context.Background()
	client = GetMongoClient()
	defer client.Disconnect(context.Background())
	file := "../examples/template.json"
	collection := "template"
	total := 100
	dbName := "keyhole"
	f := NewSeed()
	f.SetCollection(collection)
	f.SetDatabase(dbName)
	f.SetFile(file)
	f.SetIsDrop(true)
	f.SetShowProgress(false)
	f.SetTotal(total)

	// calling seedFromTemplate
	f.SeedData(client)
	db := client.Database(dbName)
	coll := db.Collection("vehicles")
	var count int64
	if count, err = coll.CountDocuments(ctx, bson.M{}); err != nil {
		t.Fatal(err)
	}

	if int64(total) != count {
		t.Fatal(errors.New("vehicles count doesn't match"))
	}
}
