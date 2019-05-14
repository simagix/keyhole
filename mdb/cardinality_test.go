// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestCheckCardinality(t *testing.T) {
	var err error
	var doc bson.M
	var client *mongo.Client
	client = getMongoClient()
	defer client.Disconnect(context.Background())

	card := NewCardinality(client)
	// card.SetVerbose(true)
	if doc, err = card.CheckCardinality(dbName, "cars"); err != nil {
		t.Fatal(err)
	}
	t.Log(Stringify(doc, "", "  "))
}
