// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Cardinality -
type Cardinality struct {
	client  *mongo.Client
	verbose bool
}

// NewCardinality returns cardinality constructor
func NewCardinality(client *mongo.Client) *Cardinality {
	return &Cardinality{client: client}
}

// SetVerbose -
func (card *Cardinality) SetVerbose(verbose bool) {
	card.verbose = verbose
}

// CheckCardinality -
func (card *Cardinality) CheckCardinality(database string, collection string) (bson.M, error) {
	var err error
	var cur *mongo.Cursor
	var ctx = context.Background()
	var doc bson.M

	if collection == "" {
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
	facetFmt := `
  [
      {"$sample": {"size": %d}},
      {"$facet": {%s}},
      {"$project": {%s}}
  ]`
	countFmt := `
	"%s": [
    {"$redact": {"$cond": {"if": { "$and": [{"$ne": [{"$type": "$%s"}, "array"]}, {"$ne": [{"$type": "$%s"}, "object"]}]},
      "then": "$$DESCEND",
      "else": "$$PRUNE"}} },
    {"$group": {"_id": "$%s"}}, {"$group": {"_id": 1,"count": {"$sum": 1}}}
	]`

	c := card.client.Database(database).Collection(collection)
	var count int64
	if count, err = c.CountDocuments(ctx, bson.M{}); err != nil {
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
	if cur, err = c.Aggregate(ctx, MongoPipeline(pipeline), opts); err != nil {
		return nil, err
	}
	if cur.Next(ctx) == false {
		cur.Close(ctx)
		return nil, err
	}
	cur.Decode(&doc)
	cur.Close(ctx)
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
	if cur, err = c.Aggregate(ctx, MongoPipeline(pipeline), opts); err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	if cur.Next(ctx) == false {
		return nil, err
	}
	doc = bson.M{}
	cur.Decode(&doc)
	for k, v := range doc {
		if fmt.Sprintf("%v", v) == "0" {
			delete(doc, k)
		}
	}
	return doc, err
}

// Print prints either JSON or tabular summary
func (card *Cardinality) Print(document bson.M) {
	if card.verbose == false {
		fmt.Println(Stringify(document, "", "   "))
	}
}
