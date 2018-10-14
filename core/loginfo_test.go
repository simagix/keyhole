// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"testing"
)

func TestGetDocByField(t *testing.T) {
	str := "2018-05-18T12:10:47.047+0000 I COMMAND  [conn3683663] command taterstore.recentlyWatched command: find { find: \"recentlyWatched\", filter: { tveUserId: 97018 }, projection: { $sortKey: { $meta: \"sortKey\" } }, sort: { updated: -1 }, shardVersion: [ Timestamp 0|0, ObjectId('000000000000000000000000') ] } planSummary: IXSCAN { tveUserId: 1, updated: -1 } cursorid:262265246743 keysExamined:101 docsExamined:101 numYields:2 nreturned:101 reslen:20674 locks:{ Global: { acquireCount: { r: 6 } }, Database: { acquireCount: { r: 3 } }, Collection: { acquireCount: { r: 3 } } } protocol:op_command 140ms"
	var matched string

	matched = getDocByField(str, "filter:")
	if matched != "{ tveUserId: 97018 }" {
		t.Fatal()
	}

	matched = getDocByField(str, "planSummary: IXSCAN")
	if matched != "{ tveUserId: 1, updated: -1 }" {
		t.Fatal()
	}
}

func TestRemoveInElements(t *testing.T) {
	str := "{ tveUserId: 12541178, titleId: { $in: [ 134734, 3425024 ] } }"
	filter := removeInElements(str, "$in: [ ")
	if filter != "{ tveUserId: 12541178, titleId: { $in: [... ] } }" {
		t.Fatal()
	}
}
