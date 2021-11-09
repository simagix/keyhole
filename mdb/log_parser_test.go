// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"regexp"
	"testing"
)

func TestParseLogFind(t *testing.T) {
	str := `2018-05-18T21:55:08.926+0000 I COMMAND  [conn3709635] command taterstore.recentlyWatched command: find { find: "recentlyWatched", filter: { tveUserId: 16114978 }, projection: { $sortKey: { $meta: "sortKey" } }, sort: { updated: -1 }, shardVersion: [ Timestamp 0|0, ObjectId('000000000000000000000000') ] } planSummary: IXSCAN { tveUserId: 1, updated: -1 } keysExamined:29 docsExamined:29 cursorExhausted:1 numYields:1 nreturned:29 reslen:6443 locks:{ Global: { acquireCount: { r: 4 } }, Database: { acquireCount: { r: 2 } }, Collection: { acquireCount: { r: 2 } } } protocol:op_command 139ms`
	loginfo := NewLogInfo("utest-xxxxxx")
	loginfo.regexp = regexp.MustCompile(loginfo.regex)
	if stat, err := loginfo.ParseLog(str); err != nil {
		t.Fatal(err)
	} else if stat.filter != `{tveUserId:1}, sort: { updated: -1 }` {
		t.Fatal(stat.filter)
	}
}

func TestParseLogWrite(t *testing.T) {
	var err error
	var stat LogStats
	str := `2021-11-08T20:59:44.939+0000 I  COMMAND  [conn188094] command testDB.$cmd command: update { update: "test2Application", bypassDocumentValidation: false, ordered: true, stmtIds: [ 0 ], updates: [ { q: { _id: "BCDDE98F-274E-4AC6-8DA8-A8E560084DE6" }, u: { _id: "BCDDE98F-274E-4AC6-8DA8-A8E560084DE6", applicationTimeStamp: new Date(1636405184882), createdDateTime: new Date(1636405184882), lastUpdateDateTime: new Date(1636405184882), serviceId: "test", version: "A0CAECC9-EFDF-42E5-BF9F-58376FBFC435", fields: { someRandom: "someRandom value of some field AD77E001-CF69-4A62-9135-1A434A667291 this is just test", simpleField: "simpleField value of some field FF958BFD-FE14-44BC-AA02-3561EF6C1C8D this is just test", lastName: "lastName value of some field 3E277AA9-98EC-4087-8608-34A8D346A841 this is just test", someField: "someField value of some field 9F52F8EC-5C1C-4F0F-9AC8-837944E529DA this is just test", testField: "testField value of some field 87B60D7B-4110-42E2-8024-600EAF71F382 this is just test", addreaField: "addreaField value of some field 002CA552-1A72-4876-BFC1-F0119E2AD44C this is just test", nextField: "nextField value of some field 6CEF4722-ED9D-4FB4-8EB4-25EEA34C6F07 this is just test", thisField: "thisField value of some field 3CD62166-C011-4678-B22D-4A4C75EF7095 this is just test", SSN: "12345678548631", firstName: "firstName value of some field 314CC8F0-90E6-4F29-A500-6D814B63CA10 this is just test", otherField: "otherField value of some field 117A15AB-0E21-451C-9697-81D9367B7091 this is just test", dob: "12-34-548631", name: "name value of some field 203084F9-A1AD-4E75-A8D2-946F7DE72DEF this is just test", somethingElse: "somethingElse value of some field 75B92310-D848-4528-AA5B-D8F50319FD58 this is just test", test2field: "test2field value of some field 1E2AF208-5A43-48EF-ADBC-BBBEE792B94A this is just test" }, _class: "com.att.bdcoe.cmt.ApplicationRecord" }, multi: false, upsert: true } ], runtimeConstants: { localNow: new Date(1636405184893), clusterTime: Timestamp(1636405184, 240) }, shardVersion: [ Timestamp(1, 14), ObjectId('6182db876f64f723fef68106') ], writeConcern: { w: "majority" }, allowImplicitCollectionCreation: false, lsid: { id: UUID("3324a1f0-8a6e-40af-9800-1a48e5504ee1"), uid: BinData(0, 8AF9A6B6AF5BB7F1B69102FA21AC1C39F1901A8DB0F266EB6247923F56C3BA8C) }, txnNumber: 35570, $clusterTime: { clusterTime: Timestamp(1636405184, 240), signature: { hash: BinData(0, 0E186E80BD2BF95781449C8C9D01B10217D4F916), keyId: 7015275607055400992 } }, $audit: { $impersonatedUsers: [ { user: "avertackprod", db: "admin" } ], $impersonatedRoles: [ { role: "dbAdminAnyDatabase", db: "admin" }, { role: "clusterMonitor", db: "admin" }, { role: "readWrite", db: "avertackDB" }, { role: "readWriteAnyDatabase", db: "admin" } ] }, $client: { driver: { name: "mongo-java-driver|sync", version: "3.11.2" }, os: { type: "Linux", name: "Linux", architecture: "amd64", version: "3.10.0-1160.21.1.el7.x86_64" }, platform: "Java/Oracle Corporation/1.8.0_221-b32", mongos: { host: "atlas-523as5-shard-01-02.gwhtw.mongodb.net:27016", client: "192.168.254.5:1062", version: "4.2.17" } }, $configServerState: { opTime: { ts: Timestamp(1636405183, 79), t: 4 } }, $db: "testDB" } numYields:0 reslen:506 locks:{ ParallelBatchWriterMode: { acquireCount: { r: 2 } }, ReplicationStateTransition: { acquireCount: { w: 3 } }, Global: { acquireCount: { w: 2 } }, Database: { acquireCount: { w: 2 } }, Collection: { acquireCount: { w: 2 } }, Mutex: { acquireCount: { r: 5 } } } flowControl:{ acquireCount: 1 } storage:{ data: { bytesRead: 24347, timeReadingMicros: 63 } } protocol:op_msg 45ms`
	loginfo := NewLogInfo("utest-xxxxxx")
	loginfo.regexp = regexp.MustCompile(loginfo.regex)
	if stat, err = loginfo.ParseLog(str); err != nil {
		t.Fatal(err)
	}
	if stat.filter != `{_id:1}` {
		t.Fatal(stat.filter)
	}
	if stat.ns != `testDB.test2Application` {
		t.Fatal(stat.ns)
	}
}
func TestRemoveInElements(t *testing.T) {
	str := "{ tveUserId: 12541178, titleId: { $in: [ 134734, 3425024 ] } }"
	filter := removeInElements(str, "$in: [ ")
	if filter != "{ tveUserId: 12541178, titleId: { $in: [... ] } }" {
		t.Fatal()
	}
}
