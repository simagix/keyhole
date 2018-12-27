// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"testing"

	"github.com/mongodb/mongo-go-driver/bson"
)

func TestGetAggregatePipeline(t *testing.T) {
	var err error
	var pipeline []bson.D
	var str = "[{\"$match\": {\"operationType\": \"update\"}}]"

	if pipeline, err = GetAggregatePipeline(str); err != nil {
		t.Fatal(err)
	}

	t.Log(pipeline)
}
