// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
)

func TestGetShards(t *testing.T) {
	var client *mongo.Client
	client = getMongoClient()
	defer client.Disconnect(context.Background())
	if _, err := GetShards(client, UnitTestURL); err != nil {
		t.Log(err)
		return
	}
}

func TestGetShardsURIList(t *testing.T) {
	var client *mongo.Client
	client = getMongoClient()
	defer client.Disconnect(context.Background())
	if _, err := GetShardsURIList(client, UnitTestURL); err != nil {
		t.Log(err)
		return
	}
}
