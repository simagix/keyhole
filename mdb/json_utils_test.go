// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"encoding/json"
	"testing"

	"github.com/simagix/gox"
	"go.mongodb.org/mongo-driver/bson"
)

func TestStringify(t *testing.T) {
	helloWorld := `{"Hello":"World"}`
	var v bson.M
	b := []byte(helloWorld)
	json.Unmarshal(b, &v)
	str := gox.Stringify(v)
	if str != helloWorld {
		t.Fatal(str, helloWorld)
	}
	str = gox.Stringify(v, "", "  ")
	t.Log(str)
}
