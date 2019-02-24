// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"testing"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
)

func TestCheckCardinality(t *testing.T) {
	var err error
	var doc bson.M
	var client *mongo.Client
	client = getMongoClient()
	defer client.Disconnect(context.Background())

	card := NewCardinality(dbName, "cars")
	// card.SetVerbose(true)
	if doc, err = card.CheckCardinality(client); err != nil {
		t.Fatal(err)
	}
	t.Log(Stringify(doc, "", "  "))
}
