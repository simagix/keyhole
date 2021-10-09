// Copyright 2018 Kuei-chun Chen. All rights reserved.

package sim

import (
	"context"
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var UnitTestURL = "mongodb://localhost/"

func getMongoClient() *mongo.Client {
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

func TestCreateIndexes(t *testing.T) {
	var docs = []bson.M{{"email": 1, "hostIp": 1}}
	connString, _ := connstring.Parse(UnitTestURL)
	runner, _ := NewRunner(connString)
	if err := runner.createIndexes(docs); err != nil {
		t.Fatal(err)
	}
}

func TestCleanup(t *testing.T) {
	var err error
	var runner *Runner

	connString, _ := connstring.Parse(UnitTestURL)
	if runner, err = NewRunner(connString); err != nil {
		t.Fatal(err)
	}
	if err = runner.Cleanup(); err != nil {
		t.Fatal(err)
	}
}
