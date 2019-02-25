// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"encoding/json"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OptimeDoc -
type OptimeDoc struct {
	T  int64 `json:"t" bson:"t"`
	TS int64 `json:"ts" bson:"ts"`
}

// GetOptime -
func GetOptime(optime interface{}) int64 {
	var ts int64
	switch optime.(type) {
	// case map[string]interface{}:
	// 	bm := optime.(map[string]interface{})
	// 	b, _ := json.Marshal(bm)
	// 	var optm OptimeDoc
	// 	json.Unmarshal(b, &optm)
	// 	ts = optm.TS >> 32
	case primitive.D:
		doc := optime.(primitive.D)
		for _, elem := range doc {
			if elem.Key == "ts" {
				b, _ := json.Marshal(elem.Value)
				var optm OptimeDoc
				json.Unmarshal(b, &optm)
				ts = int64(optm.T)
				break
			}
		}
	case primitive.Timestamp:
		b, _ := json.Marshal(optime.(primitive.Timestamp))
		var optm OptimeDoc
		json.Unmarshal(b, &optm)
		ts = int64(optm.T)
	// case []interface{}:
	// 	for _, doc := range optime.([]interface{}) {
	// 		value := (doc.(map[string]interface{}))["Value"]
	// 		b, _ := json.Marshal(value)
	// 		var optm OptimeDoc
	// 		json.Unmarshal(b, &optm)
	// 		ts = optm.T >> 32
	// 	}
	default:
		panic(fmt.Sprintf("default =>%T\n", optime))
	}

	return ts
}
