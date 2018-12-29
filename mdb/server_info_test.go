// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/mongodb/mongo-go-driver/mongo"
)

var UnitTestURL = "mongodb://localhost/"

func getMongoClient() *mongo.Client {
	var err error
	var client *mongo.Client

	if os.Getenv("DATABASE_URL") != "" {
		UnitTestURL = os.Getenv("DATABASE_URL")
	}
	if client, err = mongo.Connect(context.Background(), UnitTestURL); err != nil {
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
