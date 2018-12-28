// Copyright 2018 Kuei-chun Chen. All rights reserved.

package sim

import (
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

	if os.Getenv("DATABASE_URL") != "" {
		UnitTestURL = os.Getenv("DATABASE_URL")
	}

	if client, err = mdb.NewMongoClient(UnitTestURL); err != nil {
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
	runner, _ := NewRunner("mongodb://localhost/", "", "")
	if err := runner.CreateIndexes(docs); err != nil {
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
