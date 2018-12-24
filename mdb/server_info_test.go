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

func GetTestClient() (*mongo.Client, error) {
	var err error
	var client *mongo.Client

	if os.Getenv("DATABASE_URL") != "" {
		UnitTestURL = os.Getenv("DATABASE_URL")
	}

	if client, err = NewMongoClient(UnitTestURL, "", ""); err != nil {
		return client, err
	}

	return client, err
}

func TestGetServerInfo(t *testing.T) {
	var err error
	var client *mongo.Client
	var info ServerInfo

	uri := "mongodb://localhost/"
	if os.Getenv("DATABASE_URL") != "" {
		uri = os.Getenv("DATABASE_URL")
	}

	if client, err = NewMongoClient(uri, "", ""); err != nil {
		t.Fatal(err)
	}
	client.Connect(context.Background())
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
