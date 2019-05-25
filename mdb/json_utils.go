// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"encoding/json"
	"io/ioutil"
	"regexp"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

// Stringify return a formatted string from JSON
func Stringify(doc interface{}, opts ...string) string {
	if doc == nil {
		return ""
	}
	if len(opts) == 2 {
		b, _ := json.MarshalIndent(doc, opts[0], opts[1])
		return string(b)
	}
	b, _ := json.Marshal(doc)
	return string(b)
}

// GetDocByField get JSON string by a field
func GetDocByField(str string, field string) string {
	i := strings.Index(str, field)
	if i < 0 {
		return ""
	}
	str = strings.Trim(str[i+len(field):], " ")
	isFound := false
	bpos := 0 // begin position
	epos := 0 // end position
	for _, r := range str {
		epos++
		if isFound == false && r == '{' {
			isFound = true
			bpos++
		} else if isFound == true {
			if r == '{' {
				bpos++
			} else if r == '}' {
				bpos--
			}
		}

		if isFound == true && bpos == 0 {
			break
		}
	}
	return str[bpos:epos]
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
	str = GetDocByField(str, `"filter":`)
	re = regexp.MustCompile(`(new Date\(\S+\))`)
	str = re.ReplaceAllString(str, "\"$1\"")
	re = regexp.MustCompile(`ObjectId\(['"](\S+)['"]\)`)
	str = re.ReplaceAllString(str, "ObjectId('$1')")
	var v bson.M
	json.Unmarshal([]byte(str), &v)
	d := &Walker{}
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
						arr = append(arr, k)
					}
				}
			}
		} else if isKeyword(key) == false {
			arr = append(arr, key)
		}
	}
	return arr
}

func isKeyword(key string) bool {
	return len(key) > 0 && key[0] == '$'
}
