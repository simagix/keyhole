// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"testing"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

func TestGetOptime(t *testing.T) {
	msec := int64(1546725062)
	tm := primitive.D{}
	tm = append(tm, primitive.E{Key: "ts", Value: bson.M{"T": msec, "I": int64(10836)}})
	tm = append(tm, primitive.E{Key: "t", Value: int64(787)})
	ts := GetOptime(tm)
	if ts != msec {
		t.Fatal(msec, ts)
	}
}
