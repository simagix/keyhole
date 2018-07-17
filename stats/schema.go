// Copyright 2018 Kuei-chun Chen. All rights reserved.

package stats

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
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

type CollectionList struct {
	Cursor Cursors
}

type Cursors struct {
	FirstBatch []bson.M
}

// GetIndexes -
func GetIndexes(session *mgo.Session, dbName string, verbose bool) string {
	doc := AdminCommandOnDB(session, "listCollections", dbName)
	buf, _ := json.Marshal(doc)
	cl := CollectionList{}
	json.Unmarshal(buf, &cl)
	var buffer bytes.Buffer

	for _, coll := range cl.Cursor.FirstBatch {
		if coll["type"] == "collection" {
			buffer.WriteString(coll["name"].(string))
			buffer.WriteString("\n")
			indexes, _ := session.DB(dbName).C(coll["name"].(string)).Indexes()
			var list []string
			for _, idx := range indexes {
				list = append(list, strings.Join(idx.Key, ",")+"\tname: "+idx.Name)
			}
			sort.Strings(list)
			for _, str := range list {
				buffer.WriteString("\tkeys: ")
				buffer.WriteString(str)
				buffer.WriteString("\n")
			}
		}
	}

	return buffer.String()
}
