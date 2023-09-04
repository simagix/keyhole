// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestGetExplainSummaryReplica(t *testing.T) {
	filename := "testdata/commerceticket-replica-explain.json"
	client := getMongoClient()
	defer client.Disconnect(context.Background())
	qa := NewQueryExplainer(client)
	buffer, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	var v bson.M
	json.Unmarshal(buffer, &v)
	data, _ := bson.Marshal(v)
	bson.Unmarshal(data, &v)
	t.Log(qa.GetExplainDetails(v["explain"].(bson.M)))
}

func TestGetExplainSummaryShard(t *testing.T) {
	filename := "testdata/commerceticket-shard-explain.json"
	client := getMongoClient()
	defer client.Disconnect(context.Background())
	qa := NewQueryExplainer(client)
	buffer, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	var v bson.M
	json.Unmarshal(buffer, &v)
	data, _ := bson.Marshal(v)
	bson.Unmarshal(data, &v)
	t.Log(qa.GetExplainDetails(v["explain"].(bson.M)))
}
