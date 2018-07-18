// Copyright 2018 Kuei-chun Chen. All rights reserved.

package stats

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/simagix/keyhole/utils"
)

// CollectionsList -
type CollectionsList struct {
	Cursor CursorDoc `json:"cursor" bson:"cursor"`
}

// CursorDoc -
type CursorDoc struct {
	FirstBatch []bson.M `json:"firstBatch" bson:"firstBatch"`
}

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

// GetIndexes list all indexes of collections of databases
func GetIndexes(session *mgo.Session, dbName string, verbose bool) string {
	if dbName != "" {
		return GetIndexesFromDB(session, dbName, verbose)
	}

	var buffer bytes.Buffer
	doc := AdminCommand(session, "listDatabases")
	for _, db := range doc["databases"].([]interface{}) {
		m := db.(bson.M)
		name := m["name"].(string)
		if name == "admin" || name == "config" || name == "local" {
			continue
		}
		buffer.WriteString(GetIndexesFromDB(session, name, verbose))
	}
	return buffer.String()
}

// GetIndexesFromDB list all indexes of collections of a database
func GetIndexesFromDB(session *mgo.Session, dbName string, verbose bool) string {
	doc := AdminCommandOnDB(session, "listCollections", dbName)
	buf, _ := json.Marshal(doc)
	collectionsList := CollectionsList{}
	json.Unmarshal(buf, &collectionsList)
	var buffer bytes.Buffer

	for _, coll := range collectionsList.Cursor.FirstBatch {
		if coll["type"] == "collection" {
			buffer.WriteString("\n")
			buffer.WriteString(dbName)
			buffer.WriteString(".")
			buffer.WriteString(coll["name"].(string))
			buffer.WriteString(":\n")
			indexes, _ := session.DB(dbName).C(coll["name"].(string)).Indexes()
			var list []string
			for _, idx := range indexes {
				var strbuf bytes.Buffer
				for n, key := range idx.Key {
					if n == 0 {
						strbuf.WriteString("{ ")
					}
					strbuf.WriteString(getIndexKey(key))
					if n == len(idx.Key)-1 {
						strbuf.WriteString(" }")
					} else {
						strbuf.WriteString(", ")
					}
				}
				list = append(list, strbuf.String())
			}
			sort.Strings(list)
			for _, str := range list {
				buffer.WriteString("\t")
				buffer.WriteString(str)
				buffer.WriteString("\n")
			}
		}
	}

	return buffer.String()
}

func getIndexKey(key string) string {
	if strings.Index(key, "-") == 0 {
		return key[1:] + ": -1"
	}
	return key + ": 1"
}
