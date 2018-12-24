// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"testing"

	"github.com/mongodb/mongo-go-driver/mongo"
)

var dbName = "keyhole"

func TestGetIndexesFromDB(t *testing.T) {
	var err error
	var client *mongo.Client
	if client, err = GetTestClient(); err != nil {
		t.Fatal(err)
	}
	client.Connect(context.Background())
	defer client.Disconnect(context.Background())
	// c := client.Database(dbName).Collection("examples")
	// c.EnsureIndexKey("a")
	// c.EnsureIndexKey("a", "b", "c")
	// c.EnsureIndexKey("a", "b")
	// c.EnsureIndexKey("a", "c")
	// c.EnsureIndexKey("a", "-b")

	// get index from keyhole database
	var str string
	str = GetIndexesFromDB(client, dbName)
	t.Log(str)
}

func TestGetIndexes(t *testing.T) {
	var err error
	var client *mongo.Client
	if client, err = GetTestClient(); err != nil {
		t.Fatal(err)
	}
	client.Connect(context.Background())
	defer client.Disconnect(context.Background())
	// c := client.Database(dbName).Collection("examples")
	// c.EnsureIndexKey("a")
	// c.EnsureIndexKey("a", "b", "c")
	// c.EnsureIndexKey("a", "b")
	// c.EnsureIndexKey("a", "c")
	// c.EnsureIndexKey("a", "-b")
	// var str string
	// get all indexes
	// str = GetIndexes(client, "")
	// t.Log(str)
}
