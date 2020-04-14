// Copyright 2018 Kuei-chun Chen. All rights reserved.

package sim

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestGetSchema(t *testing.T) {
	var err error
	var client *mongo.Client
	client = getMongoClient()
	defer client.Disconnect(context.Background())

	collection := client.Database("keyhole").Collection("cars")
	filter := bson.D{{}}
	var count int64
	if count, err = collection.CountDocuments(context.Background(), filter); err != nil {
		t.Fatal(err)
	}
	if count == 0 {
		f := NewFeeder()
		f.SetTotal(10)
		f.SetIsDrop(true)
		f.SetDatabase("keyhole")
		f.SetShowProgress(false)
		f.SeedCars(client)
	}

	var str string
	if str, err = GetSchema(collection, true); err != nil {
		t.Fatal(err)
	}

	t.Log(str)
}
