// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"log"
	"testing"

	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

func TestGetShardListWithURI(t *testing.T) {
	var err error
	client := getMongoClient()
	defer client.Disconnect(context.Background())
	var shards []ShardDoc
	if shards, err = GetShards(client); err != nil {
		log.Println(err)
	}
	connString, _ := connstring.Parse(UnitTestURL)
	if _, err := GetShardListWithURI(shards, connString); err != nil {
		t.Log(err)
		return
	}
}
