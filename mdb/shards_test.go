// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"testing"

	"github.com/mongodb/mongo-go-driver/mongo"
)

func TestGetShards(t *testing.T) {
	var err error
	var list []string
	var client *mongo.Client
	if client, err = GetTestClient(); err != nil {
		t.Fatal(err)
	}

	if list, err = GetShards(client, UnitTestURL); err != nil {
		t.Log(err)
		return
	}

	t.Log(list)
}
