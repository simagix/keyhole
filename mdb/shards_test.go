// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

func TestGetAllShardURIstWithConn(t *testing.T) {
	var err error
	client := getMongoClient()
	defer client.Disconnect(context.Background())
	var shards []Shard
	if shards, err = GetShards(client); err != nil {
		t.Fatal(err)
	}
	connString, _ := connstring.Parse(UnitTestURL)
	if _, err := GetAllShardURIs(shards, connString); err != nil {
		t.Fatal(err)
	}
}

func TestGetAllShardURIs(t *testing.T) {
	uri := "mongodb+srv://user:secret@keyhole.example.com/db"
	connString, _ := connstring.Parse(uri)
	shards := []Shard{
		Shard{ID: "shard-0", Host: "shard0/shard-00-00:27017,shard-00-01:27017,shard-00-02:27017", State: 1},
		Shard{ID: "shard-1", Host: "shard1/shard-01-00:27017,shard-01-01:27017,shard-01-02:27017", State: 1},
	}
	var list []string
	var err error
	if list, err = GetAllShardURIs(shards, connString); err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatal("expected 2, but got", len(list))
	}
	expected := []string{
		"mongodb://user:secret@shard-00-00:27017,shard-00-01:27017,shard-00-02:27017/?replicaSet=shard0&authSource=admin&ssl=true",
		"mongodb://user:secret@shard-01-00:27017,shard-01-01:27017,shard-01-02:27017/?replicaSet=shard1&authSource=admin&ssl=true",
	}
	for i, v := range list {
		if v != expected[i] {
			t.Log(v)
			t.Log(expected[i])
			t.Fatal("not expected")
		}
	}
}

func TestGetAllServerURIs(t *testing.T) {
	uri := "mongodb+srv://user:secret@keyhole.example.com/db"
	connString, _ := connstring.Parse(uri)
	shards := []Shard{
		Shard{ID: "shard-0", Host: "shard0/shard-00-00:27017,shard-00-01:27017,shard-00-02:27017", State: 1},
		Shard{ID: "shard-1", Host: "shard1/shard-01-00:27017,shard-01-01:27017,shard-01-02:27017", State: 1},
	}
	var list []string
	var err error
	if list, err = GetAllServerURIs(shards, connString); err != nil {
		t.Fatal(err)
	}
	if len(list) != 6 {
		t.Fatal("expected 6, but got", len(list))
	}
	expected := []string{
		"mongodb://user:secret@shard-00-00:27017/?authSource=admin&ssl=true",
		"mongodb://user:secret@shard-00-01:27017/?authSource=admin&ssl=true",
		"mongodb://user:secret@shard-00-02:27017/?authSource=admin&ssl=true",
		"mongodb://user:secret@shard-01-00:27017/?authSource=admin&ssl=true",
		"mongodb://user:secret@shard-01-01:27017/?authSource=admin&ssl=true",
		"mongodb://user:secret@shard-01-02:27017/?authSource=admin&ssl=true",
	}
	for i, v := range list {
		if v != expected[i] {
			t.Log(v)
			t.Log(expected[i])
			t.Fatal("not expected")
		}
	}
}
