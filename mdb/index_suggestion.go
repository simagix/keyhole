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
	cnt := 0
	for _, elem := range cardList {
		if cnt < 4 {
			if contains(equalityKeys, elem.Field) {
				buffer = append(buffer, `"`+elem.Field+`": 1`)
				cnt++
			}
		} else {
			break
		}
	}

	limit := len(buffer) + 1
	for _, elem := range cardList {
		if cnt < limit {
			if contains(sortKeys, elem.Field) {
				buffer = append(buffer, `"`+elem.Field+`": 1`)
				cnt++
			}
		} else {
			break
		}
	}

	limit = len(buffer) + 2
	for _, elem := range cardList {
		if cnt < limit {
			if contains(rangeKeys, elem.Field) {
				buffer = append(buffer, `"`+elem.Field+`": 1`)
				cnt++
			}
		} else {
			break
		}
	}

	var om gox.OrderedMap
	json.Unmarshal([]byte("{ "+strings.Join(buffer, ",")+" }"), &om)
	return om
}

// GetKeys gets all fields of a odc as an array
func GetKeys(document bson.D, isrange ...bool) []string {
	filter := document.Map()
	var arr []string
	for key, val := range filter {
		if key == "$or" || key == "$and" {
			for _, elem := range val.(primitive.A) {
				for k, v := range elem.(bson.D).Map() {
					if len(k) > 0 && k[0] != '$' {
						if len(isrange) == 0 || isrange[0] == isRange(v) {
							arr = append(arr, getKey(k, v)...)
						}
					}
				}
			}
		} else if len(key) > 0 && key[0] != '$' {
			if len(isrange) == 0 || isrange[0] == isRange(val) {
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
			if k == "$gte" {
				return true
			}
		}
	}
	return false
}
