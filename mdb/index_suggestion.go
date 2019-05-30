// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"bytes"
	"encoding/json"

	"github.com/simagix/gox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetIndexSuggestion returns a recommended index by cardinalities
// index follows a principle of equality, sort, rnage
func GetIndexSuggestion(explain ExplainCommand, cardList []CardinalityCount) gox.OrderedMap {
	equalityKeys := getKeys(explain.Filter, false)
	rangeKeys := getKeys(explain.Filter, true)
	sortKeys := getKeys(explain.Sort, false)
	var buffer bytes.Buffer
	buffer.WriteString("{ ")
	cnt := 0
	for _, elem := range cardList {
		if cnt < 4 {
			if contains(equalityKeys, elem.Field) {
				if cnt > 0 {
					buffer.WriteString(", ")
				}
				buffer.WriteString(`"`)
				buffer.WriteString(elem.Field)
				buffer.WriteString(`": 1`)
				cnt++
			}
		} else {
			break
		}
	}

	for i, elem := range cardList {
		if i < 1 {
			if contains(sortKeys, elem.Field) {
				if cnt > 0 {
					buffer.WriteString(", ")
				}
				buffer.WriteString(`"`)
				buffer.WriteString(elem.Field)
				buffer.WriteString(`": 1`)
				break
			}
		} else {
			break
		}
	}

	for i, elem := range cardList {
		if i < 2 {
			if contains(rangeKeys, elem.Field) {
				if cnt > 0 {
					buffer.WriteString(", ")
				}
				buffer.WriteString(`"`)
				buffer.WriteString(elem.Field)
				buffer.WriteString(`": 1`)
				break
			}
		} else {
			break
		}
	}

	buffer.WriteString(" }")
	var om gox.OrderedMap
	json.Unmarshal(buffer.Bytes(), &om)
	return om
}

func getKeys(document bson.D, _range bool) []string {
	filter := document.Map()
	var arr []string
	for key, val := range filter {
		if key == "$or" || key == "$and" {
			for _, elem := range val.(primitive.A) {
				for k, v := range elem.(bson.D).Map() {
					if len(k) > 0 && k[0] != '$' {
						if isRange(v) == _range {
							arr = append(arr, k)
						}
					}
				}
			}
		} else if len(key) > 0 && key[0] != '$' {
			if isRange(val) == _range {
				arr = append(arr, key)
			}
		}
	}
	return arr
}

func isRange(value interface{}) bool {
	m, ok := value.(bson.D)
	if ok {
		for k := range m.Map() {
			if k == "$gte" {
				return true
			}
		}
	}
	return false
}
