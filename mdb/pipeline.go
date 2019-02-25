// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"encoding/json"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoPipeline gets aggregation pipeline from a string
func MongoPipeline(str string) mongo.Pipeline {
	var pipeline = []bson.D{}
	str = strings.TrimSpace(str)
	if strings.Index(str, "[") != 0 {
		var doc bson.M
		json.Unmarshal([]byte(str), &doc)
		var v bson.D
		b, _ := bson.Marshal(doc)
		bson.Unmarshal(b, &v)
		pipeline = append(pipeline, v)
	} else {
		var docs []bson.M
		json.Unmarshal([]byte(str), &docs)
		for _, doc := range docs {
			var v bson.D
			b, _ := bson.Marshal(doc)
			bson.Unmarshal(b, &v)
			pipeline = append(pipeline, v)
		}
	}
	return pipeline
}
