// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
)

// Cardinality -
type Cardinality struct {
	collection string
	database   string
	verbose    bool
}

// NewCardinality returns cardinality constructor
func NewCardinality(database string, collection string) *Cardinality {
	return &Cardinality{database: database, collection: collection}
}

// SetVerbose -
func (card *Cardinality) SetVerbose(verbose bool) {
	card.verbose = verbose
}

// CheckCardinality -
func (card *Cardinality) CheckCardinality(client *mongo.Client) (bson.M, error) {
	var err error
	var cur *mongo.Cursor
	var ctx = context.Background()
	var doc bson.M

	if card.collection == "" {
		return doc, errors.New("collection name is required")
	}

	keysFmt := `
  [
    {"$sample": {"size": %d}},
    {"$project":{"kvs":{"$objectToArray":"$$ROOT"}}},
    {"$unwind":"$kvs"},
    {"$group":{"_id":null,"keys":{"$addToSet":"$kvs.k"}}},
    {"$project": {"_id": 0,"keys": {"$filter": {"input": "$keys","as": "key","cond": {"$ne": ["$$key","_id"]}}}}}
  ]`

	collection := client.Database(card.database).Collection(card.collection)
	var count int64
	if count, err = collection.Count(ctx, bson.M{}); err != nil {
		return nil, err
	}

	sampleCount := count
	if sampleCount > int64(10000) { // random number
		sampleCount = int64(.0495 * float32(count))
		for sampleCount >= int64(10000) {
			sampleCount /= 10
		}
	}

	pipeline := fmt.Sprintf(keysFmt, sampleCount)
	if card.verbose {
		fmt.Println(pipeline)
	}
	opts := options.Aggregate()
	opts.SetAllowDiskUse(true)
	if cur, err = collection.Aggregate(ctx, MongoPipeline(pipeline), opts); err != nil {
		return nil, err
	}
	cur.Close(ctx)
	if cur.Next(ctx) == false {
		return nil, err
	}

	facetFmt := `
  [
      {"$sample": {"size": %d}},
      {"$facet": {%s}},
      {"$project": {%s}}
  ]
  `
	countFmt := `"%s": [
    {"$redact": {"$cond": {"if": { "$and": [{"$ne": [{"$type": "$%s"}, "array"]}, {"$ne": [{"$type": "$%s"}, "object"]}]},
      "then": "$$DESCEND",
      "else": "$$PRUNE"}} },
    {"$group": {"_id": "$%s"}}, {"$group": {"_id": 1,"count": {"$sum": 1}}}]`

	cur.Decode(&doc)
	groups := []string{}
	items := []string{}
	for _, elem := range doc["keys"].(primitive.A) {
		groups = append(groups, fmt.Sprintf(countFmt, elem.(string), elem.(string), elem.(string), elem.(string)))
		items = append(items, fmt.Sprintf("\"%s\": {\"$sum\": \"$%s.count\"}", elem.(string), elem.(string)))
	}
	pipeline = fmt.Sprintf(facetFmt, sampleCount, strings.Join(groups, ","), strings.Join(items, ","))
	if card.verbose {
		fmt.Println(pipeline)
	}
	opts = options.Aggregate()
	opts.SetAllowDiskUse(true)
	if cur, err = collection.Aggregate(ctx, MongoPipeline(pipeline), opts); err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	if cur.Next(ctx) == false {
		return nil, err
	}
	doc = bson.M{}
	cur.Decode(&doc)
	delete(doc, "keys") // a hack, the driver includes keys (a bug)
	for k, v := range doc {
		if fmt.Sprintf("%v", v) == "0" {
			delete(doc, k)
		}
	}
	return doc, err
}
