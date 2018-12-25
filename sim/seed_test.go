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

	file := ""
	collection := ""
	total := 1000
	isDrop := true
	dbName := "keyhole"

	sb := NewSeedBase(file, collection, total, isDrop, dbName)

	// calling seed()
	sb.SeedData(client)
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
	total := 1000
	isDrop := true
	dbName := "keyhole"

	sb := NewSeedBase(file, collection, total, isDrop, dbName)

	// calling seedFromTemplate
	sb.SeedData(client)
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
