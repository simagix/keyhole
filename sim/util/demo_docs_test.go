// Copyright 2018 Kuei-chun Chen. All rights reserved.

package util

import (
	"reflect"
	"testing"
)

func TestUnique(t *testing.T) {
	var arr = []string{"baseball", "tennis", "baseball", "basketball", "football"}
	var result = []string{"baseball", "tennis", "basketball"}

	arr = unique(arr, 3)
	if !reflect.DeepEqual(result, arr) {
		t.Fatal()
	}
}

func TestGetDemoDoc(t *testing.T) {
	doc := GetDemoDoc()
	if doc["email"] == "" {
		t.Fatal()
	}
}
