// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"encoding/json"
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

func TestGetServerInfo(t *testing.T) {
	var err error
	var client *mongo.Client
	var info ServerInfo
	client = getMongoClient()
	defer client.Disconnect(context.Background())
	if info, err = GetServerInfo(client); err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	bytes, _ := json.MarshalIndent(info, "", "  ")
	t.Log(string(bytes))
}

func TestListDatabaseNames(t *testing.T) {
	client := getMongoClient()
	defer client.Disconnect(context.Background())
	_, err := ListDatabaseNames(client)
	if err != nil {
		t.Fatal(err)
	}
}
