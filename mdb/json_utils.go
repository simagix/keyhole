// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"encoding/json"
	"io/ioutil"
	"regexp"

	"github.com/simagix/keyhole/sim/util"
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
func GetFilterFromFile(filename string) (map[string]interface{}, error) {
	var err error
	var doc map[string]interface{}
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
	var v map[string]interface{}
	json.Unmarshal([]byte(str), &v)
	d := &util.Walker{}
	doc = d.WalkMap(v)
	return doc, err
}
