// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
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

// GetSchemaFromCollection returns a masked first doc of a collection
func GetSchemaFromCollection(session *mgo.Session, dbName string, collection string) (string, error) {
	if collection == "" {
		return "", errors.New("usage: keyhole --schema [--file filename | --uri connection_uri --collection collection_name]")
	}
	result := bson.M{}
	c := session.DB(dbName).C(collection)
	c.Find(bson.M{}).One(&result)
	buf, _ := json.Marshal(result)
	var f interface{}
	if err := json.Unmarshal(buf, &f); err != nil {
		return "", err
	}
	doc := make(map[string]interface{})
	RandomizeDocument(&doc, f, false)
	buf, _ = json.MarshalIndent(doc, "", "   ")
	return string(buf), nil
}
