// Copyright 2018 Kuei-chun Chen. All rights reserved.

package util

import (
	"encoding/json"
	"strings"

	"github.com/mongodb/mongo-go-driver/bson"
)

// GetDemoSchema returns a demo doc
func GetDemoSchema() string {
	bytes, _ := json.MarshalIndent(GetDemoDoc(), "", "  ")
	doc := strings.Replace(string(bytes), "mongodb.", "", -1)
	doc = strings.Replace(doc, "simagix.", "", -1)
	return doc
}

// GetDemoFromFile returns a doc from a template
func GetDemoFromFile(filename string) string {
	var doc bson.M
	var err error
	if doc, err = GetDocByTemplate(filename, false); err != nil {
		return ""
	}
	buf, _ := json.MarshalIndent(doc, "", "  ")
	return string(buf)
}
