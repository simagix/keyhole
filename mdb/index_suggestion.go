// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"encoding/json"
	"strings"

	"github.com/simagix/gox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetIndexSuggestion returns a recommended index by cardinalities
// index follows a principle of equality, sort, rnage
func GetIndexSuggestion(explain ExplainCommand, cardList []CardinalityCount) gox.OrderedMap {
	equalityKeys := GetKeys(explain.Filter, false)
	rangeKeys := GetKeys(explain.Filter, true)
	sortKeys := GetKeys(explain.Sort)
	buffer := []string{}
	// the limit number is random.  Need research to support these numbers
	if contains(equalityKeys, "_id") {
		buffer = append(buffer, `"_id": 1`)
	} else {
		buffer = append(buffer, getIndexesString(cardList, equalityKeys, 4)...)
	}
	if explain.Group != "" {
		buffer = append(buffer, `"`+explain.Group+`": 1`)
	}
	buffer = append(buffer, getIndexesString(cardList, sortKeys, 1)...)
	buffer = append(buffer, getIndexesString(cardList, rangeKeys, 2)...)
	var om gox.OrderedMap
	json.Unmarshal([]byte("{ "+strings.Join(buffer, ",")+" }"), &om)
	return om
}

func getIndexesString(cardList []CardinalityCount, keys []string, max int) []string {
	buffer := []string{}
	cnt := 0
	for _, elem := range cardList {
		if contains(keys, elem.Field) {
			buffer = append(buffer, `"`+elem.Field+`": 1`)
			cnt++
		}
		if cnt >= max {
			return buffer
		}
	}
	return buffer
}

// GetKeys gets all fields of a odc as an array
func GetKeys(document bson.D, _range ...bool) []string {
	filter := document.Map()
	var arr []string
	for key, val := range filter {
		if key == "$or" || key == "$and" {
			for _, elem := range val.(primitive.A) {
				for k, v := range elem.(bson.D).Map() {
					if len(k) > 0 && k[0] != '$' {
						if len(_range) == 0 || _range[0] == isRange(v) {
							arr = append(arr, getKey(k, v)...)
						}
					}
				}
			}
		} else if len(key) > 0 && key[0] != '$' {
			if len(_range) == 0 || _range[0] == isRange(val) {
				arr = append(arr, getKey(key, val)...)
			}
		}
	}
	return arr
}

func getKey(key string, value interface{}) []string {
	strs := []string{}
	m, ok := value.(bson.D)
	if ok {
		if m.Map()["$elemMatch"] != nil {
			for k := range m.Map()["$elemMatch"].(bson.D).Map() {
				strs = append(strs, key+"."+k)
			}
			return strs
		}
	}
	strs = append(strs, key)
	return strs
}

func isRange(value interface{}) bool {
	keyMap, ok := value.(bson.D)
	if ok {
		for k := range keyMap.Map() {
			if k == "$gte" || k == "$gt" || k == "$lte" || k == "$lt" {
				return true
			}
		}
	}
	return false
}
