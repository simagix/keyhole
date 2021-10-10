// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// RunAdminCommand executes admin Command
func RunAdminCommand(client *mongo.Client, command string) (bson.M, error) {
	return RunCommandOnDB(client, command, "admin")
}

// RunCommandOnDB execute admin Command at given database
func RunCommandOnDB(client *mongo.Client, command string, db string) (bson.M, error) {
	var result = bson.M{}
	var err = client.Database(db).RunCommand(context.Background(), bson.D{{Key: command, Value: 1}}).Decode(&result)
	return result, err
}
