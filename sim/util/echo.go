// Copyright 2019 Kuei-chun Chen. All rights reserved.

package util

import (
	"encoding/json"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

// Echo echoes bson.M
func Echo(doc bson.M) {
	var b []byte
	b, _ = json.MarshalIndent(doc, "", "  ")
	fmt.Println(string(b))
}
