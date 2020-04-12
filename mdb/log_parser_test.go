// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"testing"
)

func TestParseLog(t *testing.T) {
	str := `2018-05-18T21:55:08.926+0000 I COMMAND  [conn3709635] command taterstore.recentlyWatched command: find { find: "recentlyWatched", filter: { tveUserId: 16114978 }, projection: { $sortKey: { $meta: "sortKey" } }, sort: { updated: -1 }, shardVersion: [ Timestamp 0|0, ObjectId('000000000000000000000000') ] } planSummary: IXSCAN { tveUserId: 1, updated: -1 } keysExamined:29 docsExamined:29 cursorExhausted:1 numYields:1 nreturned:29 reslen:6443 locks:{ Global: { acquireCount: { r: 4 } }, Database: { acquireCount: { r: 2 } }, Collection: { acquireCount: { r: 2 } } } protocol:op_command 139ms`
	loginfo := NewLogInfo()
	if stat, err := loginfo.ParseLog(str); err != nil {
		t.Fatal(err)
	} else if stat.filter != `{tveUserId:1}, sort: { updated: -1 }` {
		t.Fatal(stat.filter)
	}
}

func TestRemoveInElements(t *testing.T) {
	str := "{ tveUserId: 12541178, titleId: { $in: [ 134734, 3425024 ] } }"
	filter := removeInElements(str, "$in: [ ")
	if filter != "{ tveUserId: 12541178, titleId: { $in: [... ] } }" {
		t.Fatal()
	}
}
