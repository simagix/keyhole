// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"os"
	"testing"
)

func TestRemoveInElements(t *testing.T) {
	str := "{ tveUserId: 12541178, titleId: { $in: [ 134734, 3425024 ] } }"
	filter := removeInElements(str, "$in: [ ")
	if filter != "{ tveUserId: 12541178, titleId: { $in: [... ] } }" {
		t.Fatal()
	}
}

func TestLogInfo(t *testing.T) {
	loginfo := NewLogInfo()
	loginfo.SetSilent(true)
	if _, err := loginfo.Analyze("testdata/mongod.log"); err != nil {
		t.Fatal(err)
	}
	os.Remove(loginfo.OutputFilename)
}
