// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// Cardinality -
type Cardinality struct {
	client  *mongo.Client
	verbose bool
}

// CardinalitySummary stores Cardinality summary
type CardinalitySummary struct {
	SampledCount int64
	List         []CardinalityCount
}

// CardinalityCount stores cardinality counts
type CardinalityCount struct {
	Field string
	Count int64
}

// NewCardinality returns cardinality constructor
func NewCardinality(client *mongo.Client) *Cardinality {
	return &Cardinality{client: client}
}

// SetVerbose -
func (card *Cardinality) SetVerbose(verbose bool) {
	card.verbose = verbose
}

// GetCardinalityArray returns cardinality list
func (card *Cardinality) GetCardinalityArray(database string, collection string, keys ...[]string) (CardinalitySummary, error) {
	var err error
	var cur *mongo.Cursor
	var ctx = context.Background()
	var doc bson.M
	var fields []string
	summary := CardinalitySummary{}
	if collection == "" {
		return summary, errors.New("collection name is required")
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
	  {"$group": {"_id": "$%s"}}, {"$unwind": "$_id"}, {"$group": {"_id": 1,"count": {"$sum": 1}}}
	]`

	c := card.client.Database(database).Collection(collection)
	var count int64
	if count, err = c.CountDocuments(ctx, bson.M{}); err != nil {
		return summary, err
	}

	summary.SampledCount = count
	if summary.SampledCount > int64(10000) { // random number
		summary.SampledCount = int64(.0495 * float32(count))
		for summary.SampledCount >= int64(10000) {
			summary.SampledCount /= 10
		}
	}
	var pipeline string
	opts := options.Aggregate()
	if len(keys) == 0 || len(keys[0]) == 0 {
		pipeline = fmt.Sprintf(keysFmt, summary.SampledCount)
		if card.verbose {
			fmt.Println("keysFmt", pipeline)
		}
		opts.SetAllowDiskUse(true)
		if cur, err = c.Aggregate(ctx, MongoPipeline(pipeline), opts); err != nil {
			if card.verbose {
				fmt.Println("keysFmt", err)
			}
			return summary, err
		}
		if !cur.Next(ctx) {
			cur.Close(ctx)
			return summary, err
		}
		cur.Decode(&doc)
		cur.Close(ctx)
		for _, elem := range doc["keys"].(primitive.A) {
			fields = append(fields, elem.(string))
		}
	} else {
		fields = keys[0]
	}
	groups := []string{}
	items := []string{}
	for _, elem := range fields {
		groups = append(groups, fmt.Sprintf(countFmt, strings.Replace(elem, ".", "__", -1), elem))
		items = append(items, fmt.Sprintf("\"%s\": {\"$sum\": \"$%s.count\"}", strings.Replace(elem, ".", "__", -1), strings.Replace(elem, ".", "__", -1)))
	}
	pipeline = fmt.Sprintf(facetFmt, summary.SampledCount, strings.Join(groups, ","), strings.Join(items, ","))
	if card.verbose {
		fmt.Println("facetFmt", pipeline)
	}
	opts = options.Aggregate()
	opts.SetAllowDiskUse(true)
	if cur, err = c.Aggregate(ctx, MongoPipeline(pipeline), opts); err != nil {
		if card.verbose {
			fmt.Println("facetFmt", err)
		}
		return summary, err
	}
	defer cur.Close(ctx)
	if !cur.Next(ctx) {
		return summary, err
	}
	doc = bson.M{}
	cur.Decode(&doc)
	for k, v := range doc {
		if fmt.Sprintf("%v", v) == "0" {
			delete(doc, k)
		}
	}
	for k, v := range doc {
		summary.List = append(summary.List,
			CardinalityCount{Field: strings.Replace(k, "__", ".", -1), Count: toInt64(v)})
	}

	sort.Slice(summary.List, func(i, j int) bool {
		if summary.List[i].Count > summary.List[j].Count {
			return true
		} else if summary.List[i].Count == summary.List[j].Count && summary.List[i].Field < summary.List[j].Field {
			return true
		}
		return false
	})
	return summary, err
}

// GetSummary get summary of cardinality
func (card *Cardinality) GetSummary(summary CardinalitySummary) string {
	if card.verbose {
		fmt.Println("GetSummary", summary)
	}
	var buffer bytes.Buffer

	p := message.NewPrinter(language.English)
	buffer.WriteString("=> Cardinality (sampled data: " + p.Sprintf("%d", summary.SampledCount) + "):\n")
	buffer.WriteString("=========================================\n")
	for _, val := range summary.List {
		buffer.WriteString(fmt.Sprintf("%7v: %s\n", p.Sprintf("%d", val.Count), val.Field))
	}
	return buffer.String()
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
