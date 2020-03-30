// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestNewMongoClient(t *testing.T) {
	var err error

	uri := "mongodb://localhost/keyhole"
	if os.Getenv("DATABASE_URL") != "" {
		uri = os.Getenv("DATABASE_URL")
	}

	if _, err = NewMongoClient(uri); err != nil {
		t.Fatal(err)
	}
}

func TestNewMongoClientWithOptions(t *testing.T) {
	var err error
	var client *mongo.Client

	uri := "mongodb://localhost/keyhole?replicaSet=replset"
	if os.Getenv("DATABASE_URL") != "" {
		uri = os.Getenv("DATABASE_URL")
	}

	if client, err = NewMongoClient(uri, "testdata/certs/ca.pem", "testdata/certs/client.pem"); err != nil {
		t.Fatal(uri, err)
	}
	collection := client.Database("test").Collection("examples")
	var count int64
	if count, err = collection.CountDocuments(nil, bson.M{}); err != nil {
		t.Fatal(uri, err)
	}
	t.Log(count, "total counts from", uri)
}

func TestParse(t *testing.T) {
	uri := "mongodb://localhost/keyhole?replicaSet=replset"
	str, err := parse(uri)
	if err != nil || str != uri {
		t.Fatal(err)
	}
}

func TestReadPasswordFromStdin(t *testing.T) {
}
