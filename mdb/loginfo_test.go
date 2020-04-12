// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"os"
	"testing"
)

func TestLogInfo(t *testing.T) {
	loginfo := NewLogInfo()
	loginfo.SetSilent(true)
	if _, err := loginfo.Analyze("testdata/mongod.log.gz"); err != nil {
		t.Fatal(err)
	}
	os.Remove(loginfo.OutputFilename)
}
