// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"

	anly "github.com/simagix/mongo-ftdc/analytics"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// OplogStats stores oplog stats
type OplogStats struct {
	LogSizeMB int
	UsedMB    int
	TimeDiff  int
	Count     int
}

// GetOplogStats returns oplog stats
func GetOplogStats(client *mongo.Client) bson.M {
	var err error
	var cur *mongo.Cursor
	ctx := context.Background()
	oplog := bson.M{}
	db := client.Database("local")
	c := db.Collection("oplog.rs")
	var stats bson.M
	db.RunCommand(ctx, bson.D{{Key: "collStats", Value: "oplog.rs"}}).Decode(&stats)
	oplog["maxSize"] = stats["maxSize"]
	oplog["size"] = stats["size"]
	oplog["count"] = stats["count"]

	opts := options.Find()
	opts.SetProjection(bson.D{{Key: "ts", Value: 1}})
	opts.SetLimit(1)
	if cur, err = c.Find(ctx, bson.D{{}}, opts); err != nil {
		return oplog
	}
	var fdoc bson.M
	cur.Next(ctx)
	cur.Decode(&fdoc)

	opts.SetSort(bson.D{{Key: "$natural", Value: -1}})
	if cur, err = c.Find(ctx, bson.D{{}}, opts); err != nil {
		return oplog
	}
	var ldoc bson.M
	cur.Next(ctx)
	cur.Decode(&ldoc)
	oplog["durationInSeconds"] = anly.GetOptime(ldoc["ts"]) - anly.GetOptime(fdoc["ts"])
	return oplog
}
