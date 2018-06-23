// Copyright 2018 Kuei-chun Chen. All rights reserved.

package stats

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/globalsign/mgo"
	"github.com/simagix/keyhole/utils"
	"gopkg.in/mgo.v2/bson"
)

// GetDemoSchema returns a demo doc
func GetDemoSchema() string {
	bytes, _ := json.MarshalIndent(utils.GetDemoDoc(), "", "  ")
	doc := strings.Replace(string(bytes), "mongodb.", "", -1)
	doc = strings.Replace(doc, "simagix.", "", -1)
	return doc
}

// GetDemoFromFile returns a doc from a template
func GetDemoFromFile(filename string) string {
	buf, _ := json.MarshalIndent(utils.GetDocByTemplate(filename, false), "", "  ")
	return string(buf)
}

// GetSchemaFromCollection returns a masked first doc of a collection
func GetSchemaFromCollection(session *mgo.Session, dbName string, collection string, verbose bool) string {
	result := bson.M{}
	session.SetMode(mgo.Primary, true)
	c := session.DB(dbName).C(collection)
	c.Find(bson.M{}).One(&result)

	buf, _ := json.Marshal(result)
	var f interface{}
	err := json.Unmarshal(buf, &f)
	if err != nil {
		fmt.Println("Error parsing JSON: ", err)
		panic(err)
	}
	doc := make(map[string]interface{})
	utils.RandomizeDocument(&doc, f, false)
	buf, _ = json.MarshalIndent(doc, "", "   ")
	return string(buf)
}
