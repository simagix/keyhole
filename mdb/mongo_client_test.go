// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"os"
	"testing"
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

	uri := "mongodb://localhost/keyhole"
	if os.Getenv("DATABASE_URL") != "" {
		uri = os.Getenv("DATABASE_URL")
	}

	if _, err = NewMongoClient(uri, "testdata/certs/ca.pem", "testdata/certs/client.pem"); err != nil {
		t.Fatal(err)
	}
}

func TestParse(t *testing.T) {
	uri := "mongodb://localhost/keyhole?replicaSet=replset"
	str, err := Parse(uri)
	if err != nil || str != uri {
		t.Fatal(err)
	}
}

func TestReadPasswordFromStdin(t *testing.T) {
}
