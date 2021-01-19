// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"errors"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
)

func TestGetErrorCode(t *testing.T) {
	var err error
	err = mongo.WriteError{Index: 0, Code: 123, Message: "error 123"}
	if code := GetErrorCode(err); code != 123 {
		t.Fatal("expected 123 but got", code)
	}

	err = errors.New("simple error")
	if code := GetErrorCode(err); code != 0 {
		t.Fatal("expected 0 but got", code)
	}
}
