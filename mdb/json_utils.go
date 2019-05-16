// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"encoding/json"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/simagix/keyhole/sim/util"
	"go.mongodb.org/mongo-driver/bson"
)

// Silent does nothing
func Silent(data interface{}) {
}

// Stringify return a formatted string from JSON
func Stringify(doc interface{}, opts ...string) string {
	if len(opts) == 2 {
		b, _ := json.MarshalIndent(doc, opts[0], opts[1])
		return string(b)
	}
	b, _ := json.Marshal(doc)
	return string(b)
}

// GetFilterFromFile gets filter map
func GetFilterFromFile(filename string) (bson.M, error) {
	var err error
	var doc bson.M
	var buffer []byte
	if buffer, err = ioutil.ReadFile(filename); err != nil {
		return doc, err
	}
	if err = json.Unmarshal(buffer, &doc); err == nil {
		return doc, err
	}
	err = nil
	// can be a log entry
	re := regexp.MustCompile(`((\S+):)`)
	str := re.ReplaceAllString(string(buffer), "\"$2\":")
	str = util.GetDocByField(str, `"filter":`)

	re = regexp.MustCompile(`(new Date\(\S+\))`)
	str = re.ReplaceAllString(str, "\"$1\"")
	var v bson.M
	json.Unmarshal([]byte(str), &v)
	d := &util.Walker{}
	doc = d.WalkMap(v)
	return doc, err
}

// GetKeys gets all fields of a odc as an array
func GetKeys(filter bson.M) []string {
	var arr []string
	for key, val := range filter {
		if key == "$or" || key == "$and" {
			for _, elem := range val.([]interface{}) {
				for k := range elem.(map[string]interface{}) {
					if isKeyword(k) == false {
						arr = append(arr, getFirstField(k))
					}
				}
			}
		} else {
			if isKeyword(key) == false {
				arr = append(arr, getFirstField(key))
			}
		}
	}
	return arr
}

func isKeyword(key string) bool {
	// keywords := []string{"$exists", "$not", "$eq", "$ne"}
	// for _, k := range keywords {
	// 	if k == key {
	// 		return true
	// 	}
	// }
	if len(key) > 0 && key[0] == '$' {
		return true
	}
	return false
}

func getFirstField(key string) string {
	i := strings.Index(key, ".")
	if i < 0 {
		return key
	}

	return key[:i]
}
