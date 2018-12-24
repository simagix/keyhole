// Copyright 2018 Kuei-chun Chen. All rights reserved.

package util

import (
	"encoding/json"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

// CloneDoc clones a doc and assign a _id
func CloneDoc(doc bson.M) bson.M {
	_id := primitive.NewObjectID()
	var ndoc = make(bson.M)
	bytes, _ := json.Marshal(doc)
	json.Unmarshal(bytes, &ndoc)
	ndoc["_id"] = _id
	return ndoc
}
