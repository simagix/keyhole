// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestGetExplainSummaryReplica(t *testing.T) {
	filename := "testdata/commerceticket-replica-explain.json"
	client := getMongoClient()
	defer client.Disconnect(context.Background())
	qa := NewQueryAnalyzer(client)
	buffer, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	var v bson.M
	json.Unmarshal(buffer, &v)
	data, _ := bson.Marshal(v)
	bson.Unmarshal(data, &v)
	str, err := qa.GetSummary(v["explain"].(bson.M))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(str)
}

func TestGetExplainSummaryShard(t *testing.T) {
	filename := "testdata/commerceticket-shard-explain.json"
	client := getMongoClient()
	defer client.Disconnect(context.Background())
	qa := NewQueryAnalyzer(client)
	buffer, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	var v bson.M
	json.Unmarshal(buffer, &v)
	data, _ := bson.Marshal(v)
	bson.Unmarshal(data, &v)
	str, err := qa.GetSummary(v)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(str)
}
