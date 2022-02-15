// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"

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

	uri := "mongodb://localhost/keyhole?replicaSet=replset&tlsCAFile=testdata/certs/ca.pem&tlsCertificateKeyFile=testdata/certs/client.pem"
	if os.Getenv("DATABASE_URL") != "" {
		uri = os.Getenv("DATABASE_URL")
	}

	if client, err = NewMongoClient(uri); err != nil {
		t.Fatal(uri, err)
	}
	collection := client.Database("test").Collection(ExamplesCollection)
	var count int64
	if count, err = collection.CountDocuments(context.TODO(), bson.M{}); err != nil {
		t.Fatal(uri, err)
	}
	t.Log(count, "total counts from", uri)
}

func TestParseURI(t *testing.T) {
	var err error
	var cs connstring.ConnString

	uri := "mongodb://user:secret@localhost/keyhole"
	if cs, err = ParseURI(uri); err != nil {
		t.Fatal(err)
	}
	if cs.Password != "secret" {
		t.Fatal(err)
	}

	uri = "mongodb://user:$secret@localhost/keyhole"
	if cs, err = ParseURI(uri); err != nil {
		t.Fatal(err)
	}
	if cs.Password != "$secret" {
		t.Fatal(err)
	}

	uri = "mongodb://user:%5Esecret@localhost/keyhole"
	if cs, err = ParseURI(uri); err != nil {
		t.Fatal(err)
	}
	if cs.Password != "^secret" {
		t.Fatal(err)
	}

	if os.Getenv("ATLAS_URL") != "" {
		uri = os.Getenv("ATLAS_URL")
		if cs, err = ParseURI(uri); err != nil {
			t.Fatal(err)
		}
		if cs.Password != os.Getenv("ATLAS_SECRET") {
			t.Fatal(err)
		}
	}
}

func TestParseURIX509(t *testing.T) {
	uri := "mongodb+srv://demo.jgtm2.mongodb.net/?authSource=$external&authMechanism=MONGODB-X509&tlsCertificateKeyFile=testdata/certs/client.pem"
	cs, err := ParseURI(uri)
	assertEqual(t, nil, err)
	uri = cs.String() // password can be injected if missing
	_, err = NewMongoClient(uri)
	assertEqual(t, nil, err)
}

func TestParseURITLSOptions(t *testing.T) {
	var err error
	var cs connstring.ConnString

	uri := "mongodb://user:%24secret@localhost/keyhole?tlsCAFile=testdata/certs/ca.pem&tlsCertificateKeyFile=testdata/certs/client.pem"
	if cs, err = ParseURI(uri); err != nil {
		t.Fatal(err)
	}
	if cs.SSLCaFile != "testdata/certs/ca.pem" {
		t.Fatal(err)
	}
	if cs.SSLClientCertificateKeyFile != "testdata/certs/client.pem" {
		t.Fatal(err)
	}
}

func TestNewMongoClientReadPreference(t *testing.T) {
	var err error
	var client *mongo.Client

	uri := "mongodb://user:password@localhost/admin?replicaSet=replset&readPreference=secondary"
	if os.Getenv("DATABASE_URL") != "" {
		uri = os.Getenv("DATABASE_URL")
	}

	if client, err = NewMongoClient(uri); err != nil {
		t.Fatal(uri, err)
	}
	collection := client.Database("test").Collection(ExamplesCollection)
	var count int64
	if count, err = collection.CountDocuments(context.TODO(), bson.M{}); err != nil {
		t.Fatal(uri, err)
	}
	t.Log(count, "total counts from", uri)
}
