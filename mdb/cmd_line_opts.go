// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// CmdLineOpts stores build information
type CmdLineOpts struct {
	Argv   []string `bson:"argv"`
	Parsed bson.M   `bson:"parsed"`
}

// GetCmdLineOpts returns MongoDB build information
func GetCmdLineOpts(client *mongo.Client) (CmdLineOpts, error) {
	ctx := context.Background()
	var cmdLineOpts CmdLineOpts
	err := client.Database("admin").RunCommand(ctx, bson.D{{Key: "getCmdLineOpts", Value: 1}}).Decode(&cmdLineOpts)
	return cmdLineOpts, err
}
