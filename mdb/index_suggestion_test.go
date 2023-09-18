// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/simagix/gox"
	"go.mongodb.org/mongo-driver/bson"
)

func TestGetIndexSuggestionFromFilter(t *testing.T) {
	filename := "testdata/commerceticket-replica-explain.json"
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
	var explain ExplainCommand
	str := `{"filter": {"ct": "abc", "cs": {"$exists": true}}}`
	bson.UnmarshalExtJSON([]byte(str), true, &explain)
	index := GetIndexSuggestion(explain, summary.List)
	str = `{"ct":1,"cs":1}`
	if str != gox.Stringify(index) {
		t.Fatal("Expected", str, "but got", gox.Stringify(index))
	}
	t.Log(gox.Stringify(index))
}

func TestGetIndexSuggestion(t *testing.T) {
	filename := "testdata/TestGetIndexSuggestion.json"
	buffer, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	var v bson.M
	bson.UnmarshalExtJSON(buffer, true, &v)
	var summary CardinalitySummary
	data, _ := json.Marshal(v["cardinality"])
	json.Unmarshal(data, &summary)
	var explain ExplainCommand
	str := `{"filter": {"brand": "BMW", "year": {"$gt": 2017}}, "sort": {"color": 1}}`
	bson.UnmarshalExtJSON([]byte(str), true, &explain)
	index := GetIndexSuggestion(explain, summary.List)
	expected := `{"brand":1,"color":1,"year":1}`
	if gox.Stringify(index) != expected {
		t.Fatal("Expected", expected, "but got", gox.Stringify(index))
	}
	t.Log("index:", gox.Stringify(index))
}

func TestGetIndexSuggestionElemMatch(t *testing.T) {
	filename := "testdata/TestGetIndexSuggestionElemMatch.json"
	buffer, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	var v bson.M
	bson.UnmarshalExtJSON(buffer, true, &v)
	var summary CardinalitySummary
	data, _ := json.Marshal(v["cardinality"])
	json.Unmarshal(data, &summary)
	var explain ExplainCommand
	str := `{"filter": { "$and": [{ "filters": { "$elemMatch": { "k": "color", "v": "Red" } } }, { "filters": { "$elemMatch": { "k": "year", "v": { "$gt": 2017 } } } }] } }`
	bson.UnmarshalExtJSON([]byte(str), true, &explain)
	index := GetIndexSuggestion(explain, summary.List)
	expected := `{"filters.v":1,"filters.k":1}`
	if gox.Stringify(index) != expected {
		t.Fatal("Expected", expected, "but got", gox.Stringify(index))
	}
	t.Log("index:", gox.Stringify(index))
}
