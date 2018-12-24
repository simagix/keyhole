// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
)

func TestRunAdminCommand(t *testing.T) {
	var err error
	var client *mongo.Client
	var m bson.M
	if client, err = GetTestClient(); err != nil {
		t.Fatal(err)
	}
	client.Connect(context.Background())
	defer client.Disconnect(context.Background())
	if m, err = RunAdminCommand(client, "dbStats"); err != nil {
		t.Fatal(err)
	} else {
		bytes, _ := json.MarshalIndent(m, "", "  ")
		t.Log(string(bytes))
	}
}

func TestRunCommandOnDB(t *testing.T) {
	var err error
	var client *mongo.Client
	var m bson.M
	if client, err = GetTestClient(); err != nil {
		t.Fatal(err)
	}
	client.Connect(context.Background())
	defer client.Disconnect(context.Background())
	if m, err = RunCommandOnDB(client, "dbStats", "test"); err != nil {
		t.Fatal(err)
	} else {
		bytes, _ := json.MarshalIndent(m, "", "  ")
		t.Log(string(bytes))
	}
}

func TestIsMaster(t *testing.T) {
	var err error
	var client *mongo.Client
	var m bson.M
	if client, err = GetTestClient(); err != nil {
		t.Fatal(err)
	}
	client.Connect(context.Background())
	defer client.Disconnect(context.Background())
	if m, err = IsMaster(client); err != nil {
		t.Fatal(err)
	}
	t.Log(m)
}
