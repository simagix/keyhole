// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var UnitTestURL = "mongodb://localhost/?replicaSet=replset"

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

func TestListDatabaseNames(t *testing.T) {
	client := getMongoClient()
	defer client.Disconnect(context.Background())
	_, err := ListDatabaseNames(client)
	if err != nil {
		t.Fatal(err)
	}
}
