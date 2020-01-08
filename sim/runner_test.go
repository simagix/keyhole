// Copyright 2018 Kuei-chun Chen. All rights reserved.

package sim

import (
	"context"
	"os"
	"testing"

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

/*

type Runner struct {
	uri           string
	sslCAFile     string
	sslPEMKeyFile string
	tps           int
	filename      string
	verbose       bool
	peek          bool
	monitor       bool
	duration      int
	cleanup       bool
	drop          bool
	connString    connstring.ConnString
	client        *mongo.Client
}
*/
func TestCreateIndexes(t *testing.T) {
	var docs = []bson.M{bson.M{"email": 1, "hostIp": 1}}
	runner, _ := NewRunner(UnitTestURL, "", "")
	if err := runner.createIndexes(docs); err != nil {
		t.Fatal(err)
	}
}

func TestCleanup(t *testing.T) {
	var err error
	var runner *Runner

	if runner, err = NewRunner(UnitTestURL, "", ""); err != nil {
		t.Fatal(err)
	}
	if err = runner.Cleanup(); err != nil {
		t.Fatal(err)
	}
}
