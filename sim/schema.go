// Copyright 2018 Kuei-chun Chen. All rights reserved.

package sim

import (
	"context"
	"encoding/json"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/simagix/keyhole/sim/util"
)

// GetSchemaFromCollection returns a masked first doc of a collection
func GetSchemaFromCollection(client *mongo.Client, dbName string, collection string) (string, error) {
	var err error
	var result bson.M
	if collection == "" {
		return "", errors.New("usage: keyhole --schema [--file filename | --uri connection_uri --collection collection_name]")
	}
	ctx := context.Background()
	c := client.Database(dbName).Collection(collection)
	if err = c.FindOne(ctx, bson.M{}).Decode(&result); err != nil {
		return "", err
	}
	b, _ := json.Marshal(result)
	var f interface{}
	if err = json.Unmarshal(b, &f); err != nil {
		return "", err
	}
	doc := make(map[string]interface{})
	util.RandomizeDocument(&doc, f, false)
	b, _ = json.MarshalIndent(doc, "", "   ")
	return string(b), err
}
