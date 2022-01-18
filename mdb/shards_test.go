// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"testing"

	"github.com/simagix/gox"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

func TestGetAllShardURIstWithConn(t *testing.T) {
	var err error
	UnitTestURL = "mongodb://user:password@localhost/"
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
	for _, shard := range shards {
		t.Log(gox.Stringify(shard, "", "  "))
	}
}

func TestGetAllShardURIs(t *testing.T) {
	uri := "mongodb://user:secret@localhost/?authSource=admin&w=2&readPreference=secondary&tls=true"
	connString, err := connstring.Parse(uri)
	if err != nil {
		t.Fatal(err)
	}
	shards := []Shard{
		{ID: "shard-0", Host: "shard0/shard-00-00:27017,shard-00-01:27017,shard-00-02:27017", State: 1},
		{ID: "shard-1", Host: "shard1/shard-01-00:27017,shard-01-01:27017,shard-01-02:27017", State: 1},
	}
	var list []string
	if list, err = GetAllShardURIs(shards, connString); err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatal("expected 2, but got", len(list))
	}
	expected := []string{
		"mongodb://user:secret@shard-00-00:27017,shard-00-01:27017,shard-00-02:27017/?replicaSet=shard0&authSource=admin&tls=true&readPreference=secondary&w=2",
		"mongodb://user:secret@shard-01-00:27017,shard-01-01:27017,shard-01-02:27017/?replicaSet=shard1&authSource=admin&tls=true&readPreference=secondary&w=2",
	}
	for i, v := range list {
		assertEqual(t, v, expected[i])
	}
}

func TestGetAllServerURIs(t *testing.T) {
	uri := "mongodb+srv://user:secret@keyhole.example.com/db"
	connString, _ := connstring.Parse(uri)
	shards := []Shard{
		{ID: "shard-0", Host: "shard0/shard-00-00:27017,shard-00-01:27017,shard-00-02:27017", State: 1},
		{ID: "shard-1", Host: "shard1/shard-01-00:27017,shard-01-01:27017,shard-01-02:27017", State: 1},
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

func TestGetQueryParams(t *testing.T) {
	var err error
	var cs connstring.ConnString
	var expected string
	uri := "mongodb+srv://user:password@tags.jgtm2.mongodb.net/keyhole?readPreference=secondary&readPreferenceTags=nodeType:ANALYTICS"
	if cs, err = connstring.Parse(uri); err != nil {
		t.Fatal(err)
	}
	expected = "&tls=true&readPreference=secondary&readPreferenceTags=nodeType:ANALYTICS"
	assertEqual(t, expected, GetQueryParams(cs, false))

	expected = "&tls=true"
	assertEqual(t, expected, GetQueryParams(cs, true))
}

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Fatalf("%s != %s", a, b)
	}
}
