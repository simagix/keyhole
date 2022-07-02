// Copyright 2020-present Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"go.mongodb.org/mongo-driver/bson"
)

const (
	kb = 1024
	mb = kb * kb
)

// Stringify returns JSON string
func Stringify(doc interface{}) string {
	data, err := bson.MarshalExtJSON(doc, false, false)
	if err != nil {
		return ""
	}
	return string(data)
}
