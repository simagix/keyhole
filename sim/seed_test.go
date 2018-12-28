// Copyright 2018 Kuei-chun Chen. All rights reserved.

package sim

import (
	"context"
	"errors"
	"testing"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
)

func TestSeed(t *testing.T) {
	var err error
	var client *mongo.Client
	ctx := context.Background()
	if client, err = GetTestClient(); err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect(context.Background())
	dbName := "keyhole"
	total := 100
	f := NewFeeder()
	f.SetDatabase(dbName)
	f.SetTotal(total)
	f.SetIsDrop(true)
	f.SetShowProgress(false)

	// calling seed()
	f.SeedData(client)
	db := client.Database(dbName)
	coll := db.Collection("cars")
	var count int64
	if count, err = coll.Count(ctx, bson.M{}); err != nil {
		t.Fatal(err)
	}

	if int64(total) != count {
		t.Fatal(errors.New("cars count doesn't match"))
	}
}

func TestSeedFromTemplate(t *testing.T) {
	var err error
	var client *mongo.Client
	ctx := context.Background()
	if client, err = GetTestClient(); err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect(context.Background())
	file := "../examples/template.json"
	collection := "template"
	total := 100
	dbName := "keyhole"
	f := NewFeeder()
	f.SetCollection(collection)
	f.SetDatabase(dbName)
	f.SetFile(file)
	f.SetIsDrop(true)
	f.SetShowProgress(false)
	f.SetTotal(total)

	// calling seedFromTemplate
	f.SeedData(client)
	db := client.Database(dbName)
	coll := db.Collection("cars")
	var count int64
	if count, err = coll.Count(ctx, bson.M{}); err != nil {
		t.Fatal(err)
	}

	if int64(total) != count {
		t.Fatal(errors.New("cars count doesn't match"))
	}
}
