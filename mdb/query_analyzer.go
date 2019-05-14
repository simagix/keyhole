// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// QueryAnalyzer stores query analyzer info
type QueryAnalyzer struct {
	client   *mongo.Client
	database string
	verbose  bool
}

// NewQueryAnalyzer returns QueryAnalyzer
func NewQueryAnalyzer(client *mongo.Client) *QueryAnalyzer {
	return &QueryAnalyzer{client: client}
}

// SetDatabase sets database
func (qa *QueryAnalyzer) SetDatabase(database string) {
	qa.database = database
}

// SetVerbose sets verbosity
func (qa *QueryAnalyzer) SetVerbose(verbose bool) {
	qa.verbose = verbose
}

// Explain explains query plans
func (qa *QueryAnalyzer) Explain(collectionName string, filter map[string]interface{}) (bson.M, error) {
	ctx := context.Background()
	command := bson.M{"explain": bson.M{"count": collectionName, "query": filter}}
	var result = bson.M{}
	err := qa.client.Database(qa.database).RunCommand(ctx, command).Decode(&result)
	return result, err
}
