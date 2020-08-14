// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// BuildInfo stores build information
type BuildInfo struct {
	GitVersion string   `bson:"gitVersion"`
	Modules    []string `bson:"modules"`
	Version    string   `bson:"version"`
}

// GetBuildInfo returns MongoDB build information
func GetBuildInfo(client *mongo.Client) (BuildInfo, error) {
	ctx := context.Background()
	var buildInfo BuildInfo
	err := client.Database("admin").RunCommand(ctx, bson.D{{Key: "buildInfo", Value: 1}}).Decode(&buildInfo)
	return buildInfo, err
}
