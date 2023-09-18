// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestGetCardinalityArray(t *testing.T) {
	var err error
	var summary CardinalitySummary
	var client = getMongoClient()
	defer client.Disconnect(context.Background())

	card := NewCardinality(client)
	card.SetVerbose(true)
	keys := []string{"color", "style", "attribs.color", "filters.k"}
	// card.SetVerbose(true)
	if summary, err = card.GetCardinalityArray(dbName, "vehicles", keys); err != nil {
		t.Fatal(err)
	}
	t.Log(summary.List)
}

func TestGetCardinalitySummary(t *testing.T) {
	filename := "testdata/commerceticket-replica-explain.json"
	client := getMongoClient()
	defer client.Disconnect(context.Background())
	card := NewCardinality(client)
	buffer, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	var v bson.M
	json.Unmarshal(buffer, &v)
	data, _ := bson.Marshal(v)
	bson.Unmarshal(data, &v)
	var summary CardinalitySummary
	data, _ = json.Marshal(v["cardinality"])
	json.Unmarshal(data, &summary)
	t.Log(card.GetSummary(summary))
}
