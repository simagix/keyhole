// Copyright 2018 Kuei-chun Chen. All rights reserved.

package sim

import (
	"context"
	"os"
	"testing"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/simagix/keyhole/mdb"
)

var UnitTestURL = "mongodb://localhost/"

func GetTestClient() (*mongo.Client, error) {
	var err error
	var client *mongo.Client
	var ctx = context.Background()

	if os.Getenv("DATABASE_URL") != "" {
		UnitTestURL = os.Getenv("DATABASE_URL")
	}

	if client, err = mongo.NewClient(UnitTestURL); err != nil { // TODO: Add certificates
		return client, err
	}
	if err = client.Connect(ctx); err != nil {
		return client, err
	}
	return client, err
}

func TestGetShardsURIList(t *testing.T) {
	var err error
	var client *mongo.Client
	if client, err = GetTestClient(); err != nil {
		t.Fatal(err)
	}
	uriList, err := mdb.GetShardsURIList(client, UnitTestURL)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(uriList)
}

func TestCreateIndexes(t *testing.T) {
	var docs = []bson.M{bson.M{"email": 1, "hostIp": 1}}
	runner := NewRunner("mongodb://localhost/", "", "", 300, "",
		false, false, false, 1, false, false)
	runner.CreateIndexes(docs)
}

func TestCleanup(t *testing.T) {
	var err error
	var client *mongo.Client
	if client, err = GetTestClient(); err != nil {
		t.Fatal(err)
	}
	if err = Cleanup(client); err != nil {
		t.Fatal(err)
	}
}
