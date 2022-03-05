// Copyright 2018 Kuei-chun Chen. All rights reserved.

package sim

import (
	"context"
	"os"
	"testing"
)

func TestGetSchema(t *testing.T) {
	var err error
	os.Setenv("DATABASE_URL", "mongodb://user:password@localhost/")
	var client = getMongoClient()
	defer client.Disconnect(context.Background())

	collection := client.Database("keyhole").Collection("favorites")

	var str string
	if str, err = GetSchema(collection, true); err != nil {
		t.Fatal(err)
	}

	t.Log(str)
}
