// Copyright 2020-present Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestCompareMetadata(t *testing.T) {
	// {
	// 	"action": "compare_clusters",
	// 	"source_uri": "mongodb+srv://user:password@cluster.xxxxxx.mongodb.net/",
	// 	"target_uri": "mongodb+srv://user:password@cluster.xxxxxx.mongodb.net/"
	// }
	var err error
	cfg := &Config{Action: compareClusters, Signature: "TestCompareMetadata",
		SourceURI: "mongodb://user:password@localhost/",
		TargetURI: "mongodb://admin:secret@localhost:30309/"}
	if err = CompareMetadata(cfg); err != nil {
		t.Fatal(err)
	}
}

func TestDeepCompare(t *testing.T) {
	// {
	// 	"action": "compare_clusters",
	// 	"deep_compare": true,
	//  "filters": [ { "ns": "db.collection", "query": {}, "target_ns": "db.collection" } ],
	// 	"sample_size": 0,
	// 	"source_uri": "mongodb+srv://user:password@cluster.xxxxxx.mongodb.net/",
	// 	"target_uri": "mongodb+srv://user:password@cluster.xxxxxx.mongodb.net/"
	// }
	os.Remove("html/keyhole-compare.html")
	var err error
	cfg := &Config{Action: compareClusters, Signature: "TestDeepCompare",
		IsDeepCompare: true,
		Filters: []Filter{
			{NS: "sample_airbnb", Query: bson.D{}},
			{NS: "keyhole.dealers", Query: bson.D{}, TargetNS: "newns.numbers"},
			{NS: "keyhole.numbers", Query: bson.D{{Key: "a", Value: bson.D{{Key: "$lt", Value: 100}}}}},
		},
		SampleSize: 10240,
		SourceURI:  "mongodb://user:password@localhost/",
		TargetURI:  "mongodb://admin:secret@localhost:30309/"}
	var comp *Comparator
	if comp, err = NewComparator(cfg.SourceURI, cfg.TargetURI); err != nil {
		t.Fatal(err)
	}
	if err = comp.Compare(cfg.Filters, cfg.SampleSize); err != nil {
		t.Fatal(err)
	}
}
