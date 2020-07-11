// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Indexes holder indexes reader struct
type Indexes struct {
	client     *mongo.Client
	dbName     string
	filename   string
	indexesMap map[string]CollectionIndexes
	nocolor    bool
	verbose    bool
}

// AccessesDoc - accessss
type AccessesDoc struct {
	Ops   int       `json:"ops" bson:"ops"`
	Since time.Time `json:"since" bson:"since"`
}

// UsageDoc -
type UsageDoc struct {
	Host     string      `json:"host"`
	Accesses AccessesDoc `json:"accesses"`
}

// CollectionIndexes store index stats in a map
type CollectionIndexes map[string][]IndexStatsDoc

// IndexStatsDoc -
type IndexStatsDoc struct {
	Background              bool   `json:"background"`
	Collation               bson.D `json:"collation"`
	EffectiveKey            string `json:"effectiveKey"`
	ExpireAfterSeconds      int32  `json:"expireAfterSeconds"`
	Fields                  []string
	IndexKey                bson.D `json:"key"`
	IsDupped                bool   `json:"dupped"`
	IsShardKey              bool   `json:"shardKey"`
	Key                     string
	Name                    string     `json:"name"`
	PartialFilterExpression bson.D     `json:"partialFilterExpression"`
	Sparse                  bool       `json:"sparse"`
	TotalOps                int        `json:"totalOps"`
	Unique                  bool       `json:"unique"`
	Usage                   []UsageDoc `json:"stats"`
	Version                 int32      `json:"v"`
}

// NewIndexes establish seeding parameters
func NewIndexes(client *mongo.Client) *Indexes {
	gob.Register([]IndexStatsDoc{})
	hostname, _ := os.Hostname()
	return &Indexes{client: client, filename: hostname + "-index.enc", indexesMap: map[string]CollectionIndexes{}}
}

// NewIndexesReader establish seeding parameters
func NewIndexesReader(client *mongo.Client) *Indexes {
	hostname, _ := os.Hostname()
	return &Indexes{client: client, filename: hostname + "-index.enc", indexesMap: map[string]CollectionIndexes{}}
}

// SetFilename sets output file name
func (ix *Indexes) SetFilename(filename string) {
	ix.filename = strings.Replace(filename, ":", "_", -1)
}

// SetIndexesMap sets indexes map
func (ix *Indexes) SetIndexesMap(indexesMap map[string]CollectionIndexes) {
	ix.indexesMap = indexesMap
}

// SetIndexesMapFromFile File sets indexes map from a file
func (ix *Indexes) SetIndexesMapFromFile(filename string) error {
	var data []byte
	var err error
	if data, err = ioutil.ReadFile(filename); err != nil {
		return err
	}
	buffer := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buffer)
	return dec.Decode(&ix.indexesMap)
}

// SetNoColor set nocolor flag
func (ix *Indexes) SetNoColor(nocolor bool) {
	ix.nocolor = nocolor
}

// SetVerbose sets verbose level
func (ix *Indexes) SetVerbose(verbose bool) {
	ix.verbose = verbose
}

// SetDBName sets verbose level
func (ix *Indexes) SetDBName(dbName string) {
	ix.dbName = dbName
}

// GetIndexes list all indexes of collections of databases
func (ix *Indexes) GetIndexes() (map[string]CollectionIndexes, error) {
	var err error
	var dbNames []string
	ix.indexesMap = map[string]CollectionIndexes{} // reset
	if ix.dbName != "" {
		ix.indexesMap[ix.dbName], err = ix.GetIndexesFromDB(ix.dbName)
		return ix.indexesMap, err
	}

	if dbNames, err = ListDatabaseNames(ix.client); err != nil {
		return ix.indexesMap, err
	}
	cnt := 0
	for _, name := range dbNames {
		if name == "admin" || name == "config" || name == "local" {
			if ix.verbose == true {
				log.Println("Skip", name)
			}
			continue
		}
		cnt++
		if ix.verbose == true {
			log.Println("checking", name)
		}
		if ix.indexesMap[name], err = ix.GetIndexesFromDB(name); err != nil {
			return ix.indexesMap, err
		}
	}
	if cnt == 0 && ix.verbose == true {
		log.Println("No database is available")
	}
	return ix.indexesMap, err
}

// GetIndexesFromDB list all indexes of collections of a database
func (ix *Indexes) GetIndexesFromDB(dbName string) (CollectionIndexes, error) {
	var err error
	var cur *mongo.Cursor
	var ctx = context.Background()
	var indexesMap = CollectionIndexes{}
	if cur, err = ix.client.Database(dbName).ListCollections(ctx, bson.M{}); err != nil {
		return indexesMap, err
	}
	defer cur.Close(ctx)
	collections := []string{}
	for cur.Next(ctx) {
		var elem = map[string]interface{}{}
		if err = cur.Decode(&elem); err != nil {
			if ix.verbose == true {
				log.Println(err)
			}
			continue
		}
		coll := fmt.Sprintf("%v", elem["name"])
		collType := fmt.Sprintf("%v", elem["type"])
		if strings.Index(coll, "system.") == 0 || (elem["type"] != nil && collType != "collection") {
			if ix.verbose == true {
				log.Println("skip", coll)
			}
			continue
		}
		collections = append(collections, coll)
	}

	sort.Strings(collections)
	for _, collection := range collections {
		indexesMap[collection] = ix.GetIndexesFromCollection(ix.client.Database(dbName).Collection(collection))
	}
	return indexesMap, err
}

// GetIndexesFromCollection gets indexes from a collection
func (ix *Indexes) GetIndexesFromCollection(collection *mongo.Collection) []IndexStatsDoc {
	var err error
	var ctx = context.Background()
	var pipeline = MongoPipeline(`{"$indexStats": {}}`)
	var list []IndexStatsDoc
	var icur *mongo.Cursor
	var scur *mongo.Cursor
	if ix.verbose {
		log.Println("* GetIndexesFromCollection")
	}

	if scur, err = collection.Aggregate(ctx, pipeline); err != nil {
		log.Println(err)
		return list
	}
	var indexStats = []map[string]interface{}{}
	for scur.Next(ctx) {
		var result = map[string]interface{}{}
		if err = scur.Decode(&result); err != nil {
			log.Println(err)
			continue
		}
		indexStats = append(indexStats, result)
	}
	scur.Close(ctx)

	db := collection.Database().Name()
	cmd := bson.D{{Key: "listIndexes", Value: collection.Name()}}
	if icur, err = ix.client.Database(db).RunCommandCursor(ctx, cmd); err != nil {
		log.Println(err)
		return list
	}
	defer icur.Close(ctx)

	for icur.Next(ctx) {
		var idx = bson.D{}
		if err = icur.Decode(&idx); err != nil {
			log.Println(err)
			continue
		}

		var indexName string
		o := IndexStatsDoc{}
		for _, v := range idx {
			if v.Key == "name" {
				indexName = v.Value.(string)
				o.Name = indexName
			} else if v.Key == "key" {
				o.IndexKey = v.Value.(bson.D)
			} else if v.Key == "background" {
				o.Background, _ = v.Value.(bool)
			} else if v.Key == "expireAfterSeconds" {
				o.ExpireAfterSeconds = toInt32(v.Value)
			} else if v.Key == "sparse" {
				o.Sparse = v.Value.(bool)
			} else if v.Key == "unique" {
				o.Unique = v.Value.(bool)
			} else if v.Key == "partialFilterExpression" {
				o.PartialFilterExpression = v.Value.(bson.D)
			} else if v.Key == "collation" {
				o.Collation = v.Value.(bson.D)
			} else if v.Key == "v" {
				o.Version = v.Value.(int32)
			} else if v.Key == "ns" {
			} else if ix.verbose == true {
				fmt.Println("additional attrib", v.Key, v.Value)
			}
		}

		var strbuf bytes.Buffer
		fields := []string{}
		for n, value := range o.IndexKey {
			fields = append(fields, value.Key)
			if n == 0 {
				strbuf.WriteString("{ ")
			}
			strbuf.WriteString(value.Key + ": " + fmt.Sprint(value.Value))
			if n == len(o.IndexKey)-1 {
				strbuf.WriteString(" }")
			} else {
				strbuf.WriteString(", ")
			}
		}
		o.Fields = fields
		o.Key = strbuf.String()
		// Check shard keys
		var v map[string]interface{}
		ns := collection.Database().Name() + "." + collection.Name()
		if ix.verbose {
			log.Println("GetIndexesFromCollection", ns, o.Key)
		}
		if err = ix.client.Database("config").Collection("collections").FindOne(ctx, bson.M{"_id": ns, "key": o.IndexKey}).Decode(&v); err == nil {
			o.IsShardKey = true
		}
		o.EffectiveKey = strings.Replace(o.Key[2:len(o.Key)-2], ": -1", ": 1", -1)
		o.Usage = []UsageDoc{}
		for _, result := range indexStats {
			if result["name"].(string) == indexName {
				b, _ := bson.Marshal(result)
				var usage UsageDoc
				bson.Unmarshal(b, &usage)
				o.TotalOps += usage.Accesses.Ops
				o.Usage = append(o.Usage, usage)
			}
		}
		list = append(list, o)
	}
	sort.Slice(list, func(i, j int) bool { return (list[i].EffectiveKey < list[j].EffectiveKey) })
	for i, o := range list {
		if o.Key != "{ _id: 1 }" && o.IsShardKey == false {
			list[i].IsDupped = checkIfDupped(o, list)
		}
	}
	return list
}

// check if an index is a dup of others
func checkIfDupped(doc IndexStatsDoc, list []IndexStatsDoc) bool {
	for _, o := range list {
		// check indexes if not marked as dupped, has the same first field, and more or equal number of fields
		if o.IsDupped == false && doc.Fields[0] == o.Fields[0] && doc.Key != o.Key && len(o.Fields) >= len(doc.Fields) {
			nmatched := 0
			for i, fld := range doc.Fields {
				if i == 0 {
					continue
				}
				for j, field := range o.Fields {
					if j > 0 && fld == field {
						nmatched++
						break
					}
				}
			}
			if nmatched == len(doc.Fields)-1 {
				return true
			}
		}
	}
	return false
}

// Print prints indexes
func (ix *Indexes) Print(indexesMap map[string]CollectionIndexes) {
	var dbkeys []string
	for k := range indexesMap {
		dbkeys = append(dbkeys, k)
	}
	sort.Strings(dbkeys)
	for _, key := range dbkeys {
		collectionIndexes := indexesMap[key]
		var keys []string
		for k := range collectionIndexes {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			list := collectionIndexes[k]
			var buffer bytes.Buffer
			ns := key + "." + k
			buffer.WriteString("\n")
			buffer.WriteString(ns)
			buffer.WriteString(":\n")
			for _, o := range list {
				font := codeDefault
				tailCode := codeDefault
				if ix.nocolor {
					font = ""
					tailCode = ""
				}
				if o.Key == "{ _id: 1 }" {
					buffer.WriteString(fmt.Sprintf("%v  %v%v", font, o.Key, tailCode))
				} else if o.IsShardKey == true {
					buffer.WriteString(fmt.Sprintf("%v* %v%v", font, o.Key, tailCode))
				} else if o.IsDupped == true {
					if ix.nocolor == false {
						font = codeRed
					}
					buffer.WriteString(fmt.Sprintf("%vx %v%v", font, o.Key, tailCode))
				} else if o.TotalOps == 0 {
					if ix.nocolor == false {
						font = codeBlue
					}
					buffer.WriteString(fmt.Sprintf("%v? %v%v", font, o.Key, tailCode))
				} else {
					buffer.WriteString(fmt.Sprintf("  %v", o.Key))
				}

				for _, u := range o.Usage {
					buffer.Write([]byte("\n\thost: " + u.Host + ", ops: " + fmt.Sprintf("%v", u.Accesses.Ops) + ", since: " + fmt.Sprintf("%v", u.Accesses.Since)))
				}
				buffer.WriteString("\n")
			}
			fmt.Println(buffer.String())
		}
	}
}

// CreateIndexes creates indexes
func (ix *Indexes) CreateIndexes() error {
	var ctx = context.Background()
	var err error
	for db := range ix.indexesMap {
		indexes := ix.indexesMap[db]
		for k, list := range indexes {
			collection := ix.client.Database(db).Collection(k)
			for _, o := range list {
				if o.IsShardKey == true {
					// TODO
				}
				var indexKey bson.D
				for _, field := range o.Fields {
					for _, e := range o.IndexKey {
						if field == e.Key {
							indexKey = append(indexKey, e)
							break
						}
					}
				}

				opt := options.Index()
				opt.SetVersion(o.Version)
				opt.SetName(o.Name)
				if o.Background == true {
					opt.SetBackground(o.Background)
				}
				if o.ExpireAfterSeconds > 0 {
					opt.SetExpireAfterSeconds(o.ExpireAfterSeconds)
				}
				if o.Unique == true {
					opt.SetUnique(o.Unique)
				}
				if o.Sparse == true {
					opt.SetSparse(o.Sparse)
				}
				if o.Collation != nil {
					var collation *options.Collation
					if data, err := bson.Marshal(o.Collation); err != nil {
						fmt.Println(err)
					} else {
						bson.Unmarshal(data, &collation)
						opt.SetCollation(collation)
					}
				}
				if o.PartialFilterExpression != nil {
					opt.SetPartialFilterExpression(o.PartialFilterExpression)
				}
				if _, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: o.IndexKey, Options: opt}); err != nil {
					fmt.Println(err)
				}
			}
		}
	}
	return err
}

// Save saves indexes map to a file
func (ix *Indexes) Save() error {
	var data bytes.Buffer
	var err error
	enc := gob.NewEncoder(&data)
	if err = enc.Encode(ix.indexesMap); err != nil {
		return err
	}
	fmt.Println("Encoded indexes info is written to", ix.filename)
	return ioutil.WriteFile(ix.filename, data.Bytes(), 0644)
}
