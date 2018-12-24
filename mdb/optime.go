// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"encoding/json"
	"fmt"

	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

// OptimeDoc -
type OptimeDoc struct {
	T  int   `json:"t" bson:"t"`
	TS int64 `json:"ts" bson:"ts"`
}

// GetOptime -
func GetOptime(optime interface{}) int64 {
	var ts int64
	switch optime.(type) {
	case map[string]interface{}:
		bm := optime.(map[string]interface{})
		b, _ := json.Marshal(bm)
		var optm OptimeDoc
		json.Unmarshal(b, &optm)
		ts = optm.TS >> 32
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
	default:
		panic(fmt.Sprintf("default =>%T\n", optime))
	}

	return ts
}
