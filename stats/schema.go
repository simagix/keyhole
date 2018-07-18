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
	dbNames, _ := session.DatabaseNames()
	for _, name := range dbNames {
		if name == "admin" || name == "config" || name == "local" {
			continue
		}
		buffer.WriteString(GetIndexesFromDB(session, name, verbose))
	}
	return buffer.String()
}

// GetIndexesFromDB list all indexes of collections of a database
func GetIndexesFromDB(session *mgo.Session, dbName string, verbose bool) string {
	var buffer bytes.Buffer
	collNames, _ := session.DB(dbName).CollectionNames()

	for _, coll := range collNames {
		if strings.Index(coll, "system.") == 0 {
			continue
		}
		indexes, _ := session.DB(dbName).C(coll).Indexes()
		if len(indexes) == 1 && !verbose {
			continue
		}
		buffer.WriteString("\n")
		buffer.WriteString(dbName)
		buffer.WriteString(".")
		buffer.WriteString(coll)
		buffer.WriteString(":\n")
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
			if str == "{ _id: 1 }" && !verbose {
				continue
			}
			buffer.WriteString("\t")
			buffer.WriteString(str)
			buffer.WriteString("\n")
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
