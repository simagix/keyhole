// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"encoding/json"

	"github.com/mongodb/mongo-go-driver/bson"
)

// GetAggregatePipeline get aggregation pipeline from a string
func GetAggregatePipeline(str string) ([]bson.D, error) {
	var err error
	var pipeline = []bson.D{}
	var docs []bson.M
	json.Unmarshal([]byte(str), &docs)
	for _, doc := range docs {
		var v bson.D
		b, _ := bson.Marshal(doc)
		bson.Unmarshal(b, &v)
		pipeline = append(pipeline, v)
	}
	return pipeline, err
}
