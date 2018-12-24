// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"time"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
)

// RunAdminCommand executes admin Command
func RunAdminCommand(client *mongo.Client, command string) (bson.M, error) {
	return RunCommandOnDB(client, command, "admin")
}

// RunCommandOnDB execute admin Command at given database
func RunCommandOnDB(client *mongo.Client, command string, db string) (bson.M, error) {
	var result = bson.M{}
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	singleResult := client.Database(db).RunCommand(ctx, bson.D{{Key: command, Value: 1}})
	err = singleResult.Decode(&result)
	return result, err
}

// IsMaster executes dbisMaster()
func IsMaster(client *mongo.Client) (bson.M, error) {
	return RunAdminCommand(client, "isMaster")
}
