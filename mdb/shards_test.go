// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"testing"

	"github.com/mongodb/mongo-go-driver/mongo"
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
