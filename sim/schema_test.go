// Copyright 2018 Kuei-chun Chen. All rights reserved.

package sim

import (
	"context"
	"testing"
)

func TestGetSchema(t *testing.T) {
	var err error
	var client = getMongoClient()
	defer client.Disconnect(context.Background())

	collection := client.Database("keyhole").Collection("vehicles")

	var str string
	if str, err = GetSchema(collection, true); err != nil {
		t.Fatal(err)
	}

	t.Log(str)
}
