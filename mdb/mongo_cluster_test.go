// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func getClient() *mongo.Client {
	var err error
	var client *mongo.Client

	if os.Getenv("DATABASE_URL") != "" {
		UnitTestURL = os.Getenv("DATABASE_URL")
	}
	if client, err = mongo.Connect(context.Background(), options.Client().ApplyURI(UnitTestURL)); err != nil {
		panic(err)
	}

	return client
}

func TestGetClusterInfo(t *testing.T) {
	client := getClient()
	mc := NewMongoCluster(client)
	mc.SetVerbose(true)
	cluster, err := mc.GetClusterInfo()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(Stringify(cluster, "", "  "))
}
