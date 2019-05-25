// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"reflect"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Walker is an empty JSON document
type Walker struct{}

// WalkMap walks a map
func (walker *Walker) WalkMap(docMap map[string]interface{}) map[string]interface{} {
	for k, v := range docMap {
		vt := reflect.TypeOf(v)
		switch vt.Kind() {
		case reflect.Map:
			if mv, ok := v.(map[string]interface{}); ok {
				docMap[k] = walker.WalkMap(mv)
			} else {
				panic(v)
			}
		case reflect.Array, reflect.Slice:
			if mv, ok := v.([]interface{}); ok {
				docMap[k] = walker.WalkArray(mv)
			} else {
				panic(v)
			}
		default:
			docMap[k] = walker.convert(v)
		}
	}
	return docMap
}

// WalkArray walks an array
func (walker *Walker) WalkArray(arrayType []interface{}) []interface{} {
	for k, v := range arrayType {
		vt := reflect.TypeOf(v)
		switch vt.Kind() {
		case reflect.Map:
			if mv, ok := v.(map[string]interface{}); ok {
				arrayType[k] = walker.WalkMap(mv)
			} else {
				panic(v)
			}
		case reflect.Array, reflect.Slice:
			if mv, ok := v.([]interface{}); ok {
				arrayType[k] = walker.WalkArray(mv)
			} else {
				panic(v)
			}
		default:
			arrayType[k] = walker.convert(v)
		}

	}
	return arrayType
}

func (walker *Walker) convert(x interface{}) interface{} {
	if v, ok := x.(string); ok {
		if strings.HasPrefix(v, "new Date(") {
			ms, _ := strconv.ParseInt(v[9:len(v)-1], 10, 64)
			return time.Unix(0, ms*int64(time.Millisecond))
		} else if strings.HasPrefix(v, "ObjectId(") {
			// expect ObjectId('<hex>')
			_id, _ := primitive.ObjectIDFromHex(v[10 : len(v)-2])
			return _id
		}
	}
	return x
}
