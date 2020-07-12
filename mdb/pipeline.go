// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoPipeline gets aggregation pipeline from a string
func MongoPipeline(str string) mongo.Pipeline {
	var pipeline = []bson.D{}
	str = strings.TrimSpace(str)
	if strings.Index(str, "[") != 0 {
		var doc bson.D
		bson.UnmarshalExtJSON([]byte(str), false, &doc)
		pipeline = append(pipeline, doc)
	} else {
		bson.UnmarshalExtJSON([]byte(str), false, &pipeline)
	}
	return pipeline
}
