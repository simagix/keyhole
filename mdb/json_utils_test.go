// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"encoding/json"
	"testing"

	"github.com/mongodb/mongo-go-driver/bson"
)

func TestSilent(t *testing.T) {
	Silent(bson.M{"Hello": "World"})
}

func TestStringify(t *testing.T) {
	helloWorld := `{"Hello":"World"}`
	var v bson.M
	b := []byte(helloWorld)
	json.Unmarshal(b, &v)
	str := Stringify(v)
	if str != helloWorld {
		t.Fatal(str, helloWorld)
	}
	str = Stringify(v, "", "  ")
	t.Log(str)
}
