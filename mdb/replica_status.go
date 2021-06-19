// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ReplSetGetStatus stores replset info
type ReplSetGetStatus struct {
	Members []struct {
		ID             int                              `bson:"_id"`
		Health         int                              `bson:"health"`
		Name           string                           `bson:"name"`
		Optime         struct{ TS primitive.Timestamp } `bson:"optime"`
		State          int                              `bson:"state"`
		StateStr       string                           `bson:"stateStr"`
		SyncingTo      string                           `bson:"syncingTo"`
		SyncSourceHost string                           `bson:"syncSourceHost"`
		Uptime         int64                            `bson:"uptime"`
	} `bson:"members"`
	Set            string `bson:"set"`
	SyncingTo      string `bson:"syncingTo"`
	SyncSourceHost string `bson:"syncSourceHost"`
}

// OplogStats stores oplog stats
type OplogStats struct {
	Count             int   `bson:"count"`
	DurationInSeconds int64 `bson:"durationInSeconds"`
	MaxSize           int64 `bson:"maxSize"`
	Size              int64 `bson:"size"`
}

// GetReplSetGetStatus returns MongoDB build information
func GetReplSetGetStatus(client *mongo.Client) (ReplSetGetStatus, error) {
	ctx := context.Background()
	var replSetGetStatus ReplSetGetStatus
	err := client.Database("admin").RunCommand(ctx, bson.D{{Key: "replSetGetStatus", Value: 1}}).Decode(&replSetGetStatus)
	return replSetGetStatus, err
}

// GetOplogStats returns oplog stats
func GetOplogStats(client *mongo.Client) (OplogStats, error) {
	var err error
	var cur *mongo.Cursor
	ctx := context.Background()
	oplog := OplogStats{}
	db := client.Database("local")
	db.RunCommand(ctx, bson.D{{Key: "collStats", Value: "oplog.rs"}}).Decode(&oplog)

	c := db.Collection("oplog.rs")
	opts := options.Find()
	opts.SetProjection(bson.D{{Key: "ts", Value: 1}})
	opts.SetLimit(1)
	if cur, err = c.Find(ctx, bson.D{{}}, opts); err != nil {
		return oplog, err
	}
	var first struct{ TS primitive.Timestamp }
	cur.Next(ctx)
	cur.Decode(&first)

	opts.SetSort(bson.D{{Key: "$natural", Value: -1}})
	if cur, err = c.Find(ctx, bson.D{{}}, opts); err != nil {
		return oplog, err
	}
	var lastest struct{ TS primitive.Timestamp }
	cur.Next(ctx)
	cur.Decode(&lastest)
	oplog.DurationInSeconds = int64(lastest.TS.T - first.TS.T)
	return oplog, err
}
