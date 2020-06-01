// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"bufio"
	"os"
	"strings"
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

func TestLoginfoParse(t *testing.T) {
	str := `2020-05-12T12:48:14.398-0400 I COMMAND [conn6800] command arlmd01p.defaultCollectionAccount command: getMore { getMore: 78661681062, collection: "defaultCollectionAccount", $db: "arlmd01p" } originatingCommand: { find: "defaultCollectionAccount", filter: { $and: [ { dataLoadDate: new Date(1589169600000) }, { sourceSystem: "MSP" } ] }, sort: {}, skip: 0, $readPreference: { mode: "secondaryPreferred" }, $db: "arlmd01p" } planSummary: COLLSCAN cursorid:78661681062 keysExamined:0 docsExamined:11485 numYields:133 nreturned:2498 reslen:16771175 locks:{ Global: { acquireCount: { r: 268 } }, Database: { acquireCount: { r: 134 } }, Collection: { acquireCount: { r: 134 } } } protocol:op_query 2352ms`
	loginfo := NewLogInfo()
	loginfo.SetSilent(true)
	loginfo.SetVerbose(true)
	rd := bufio.NewReader(strings.NewReader(str))
	if err := loginfo.Parse(rd); err != nil {
		t.Fatal(err)
	}
}
