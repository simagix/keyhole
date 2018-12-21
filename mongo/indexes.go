// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mongo

import (
	"bytes"
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"github.com/globalsign/mgo"
	"github.com/mongodb/mongo-go-driver/bson"
)

// UsageDoc -
type UsageDoc struct {
	Hostname string
	Ops      int
	Since    string
}

// IndexStatsDoc -
type IndexStatsDoc struct {
	Key          string
	EffectiveKey string
	Usage        []UsageDoc
}

// GetIndexes list all indexes of collections of databases
func GetIndexes(session *mgo.Session, dbName string) string {
	if dbName != "" {
		return GetIndexesFromDB(session, dbName)
	}

	var buffer bytes.Buffer
	dbNames, _ := session.DatabaseNames()
	for _, name := range dbNames {
		if name == "admin" || name == "config" || name == "local" {
			continue
		}
		buffer.WriteString(GetIndexesFromDB(session, name))
	}
	return buffer.String()
}

// GetIndexesFromDB list all indexes of collections of a database
func GetIndexesFromDB(session *mgo.Session, dbName string) string {
	var buffer bytes.Buffer
	collNames, _ := session.DB(dbName).CollectionNames()
	pipeline := [1]bson.M{bson.M{"$indexStats": bson.M{}}}

	for _, coll := range collNames {
		if strings.Index(coll, "system.") == 0 {
			continue
		}
		results := []bson.M{}
		err := session.DB(dbName).C(coll).Pipe(pipeline).All(&results)
		if err != nil || len(results) == 0 {
			continue
		}
		indexes, _ := session.DB(dbName).C(coll).Indexes()

		buffer.WriteString("\n")
		buffer.WriteString(dbName)
		buffer.WriteString(".")
		buffer.WriteString(coll)
		buffer.WriteString(":\n")
		var list []IndexStatsDoc

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

			o := IndexStatsDoc{Key: strbuf.String()}
			o.EffectiveKey = strings.Replace(o.Key[:len(o.Key)-2], ": -1", ": 1", -1)
			o.Usage = []UsageDoc{}
			for _, result := range results {
				if result["name"].(string) == idx.Name {
					accesses := result["accesses"].(bson.M)
					host := result["host"].(string)
					ops, _ := json.Marshal(accesses["ops"])
					since, _ := json.Marshal(accesses["since"])
					x, _ := strconv.Atoi(string(ops))
					u := UsageDoc{host, x, string(since)}
					o.Usage = append(o.Usage, u)
				}
			}
			list = append(list, o)
		}

		sort.Slice(list, func(i, j int) bool { return list[i].EffectiveKey < list[j].EffectiveKey })

		for i, o := range list {
			font := "\x1b[0m  "
			if o.Key != "{ _id: 1 }" {
				if i < len(list)-1 && strings.Index(list[i+1].EffectiveKey, o.EffectiveKey) == 0 {
					font = "\x1b[31;1mx " // red
				} else {
					sum := 0
					for _, u := range o.Usage {
						sum += u.Ops
					}
					if sum == 0 {
						font = "\x1b[34;1m? " // blue
					}
				}
			}
			buffer.WriteString(font + o.Key + "\x1b[0m")
			for _, u := range o.Usage {
				buffer.Write([]byte("\n\thost: " + u.Hostname + ", ops: " + strconv.Itoa(u.Ops) + ", since: " + u.Since))
			}
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
