// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"errors"
	"testing"

	"github.com/mongodb/mongo-go-driver/bson"
)

func TestGetAggregatePipeline(t *testing.T) {
	var pipeline []bson.D
	var str = `[{"$match": {"operationType": "update"}}]`
	if pipeline = GetAggregatePipeline(str); len(pipeline) == 0 {
		t.Fatal(errors.New("empty pipeline"))
	}
	t.Log(pipeline)

	str = `[ {"$match": { "color": "Red" }},
		{"$group": { "_id": "$brand", "count": { "$sum": 1 } }},
		{"$project": { "brand": "$_id", "_id": 0, "count": 1 }} ]`
	if pipeline = GetAggregatePipeline(str); len(pipeline) == 0 {
		t.Fatal(errors.New("empty pipeline"))
	}
	t.Log(pipeline)
}
