// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"reflect"
	"testing"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

func TestCloneDoc(t *testing.T) {
	var edoc = bson.M{"name": "keyhole"}
	var doc = bson.M{"_id": primitive.NewObjectID(), "sub": edoc}
	newDoc := cloneDoc(doc)
	if reflect.DeepEqual(doc["sub"], newDoc["sub"]) == true {
		t.Fatal()
	}
}

func TestCreateIndexes(t *testing.T) {
}
