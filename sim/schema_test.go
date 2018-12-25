// Copyright 2018 Kuei-chun Chen. All rights reserved.

package sim

import (
	"context"
	"testing"

	"github.com/mongodb/mongo-go-driver/mongo"
)

func TestGetSchemaFromCollection(t *testing.T) {
	var err error
	var client *mongo.Client
	if client, err = GetTestClient(); err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	var str string
	if str, err = GetSchemaFromCollection(client, "keyhole", "cars"); err != nil {
		t.Fatal(err)
	}

	t.Log(str)
}
