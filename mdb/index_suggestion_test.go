// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/simagix/gox"
	"go.mongodb.org/mongo-driver/bson"
)

func TestGetIndexSuggestionFromFilter(t *testing.T) {
	filename := "testdata/commerceticket-replica-explain.json"
	buffer, err := ioutil.ReadFile(filename)
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
	if `{"ct":1,"cs":1}` != gox.Stringify(index) {
		t.Fatal("Expected", `{ "ct": 1, "cs": 1 }`, "but got", gox.Stringify(index))
	}
	t.Log(gox.Stringify(index))
}

func TestGetIndexSuggestion(t *testing.T) {
	filename := "testdata/cars-explain.json"
	buffer, err := ioutil.ReadFile(filename)
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
	expected := `{"year":1,"brand":1,"color":1}`
	if gox.Stringify(index) != expected {
		t.Fatal("Expected", expected, "but got", gox.Stringify(index))
	}
	// t.Log("filter:", gox.Stringify(explain.Filter.Map()))
	// t.Log("sort:", gox.Stringify(explain.Sort.Map()))
	t.Log("index:", gox.Stringify(index))
}
