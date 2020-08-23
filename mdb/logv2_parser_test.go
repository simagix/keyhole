// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func TestParseLogv2(t *testing.T) {
	str := `{"t":{"$date":"2020-03-30T09:18:46.696-04:00"},"s":"I", "c":"-","id":51803,"ctx":"conn487","msg":"slow query","attr":{"type":"remove","ns":"_KEYHOLE_88800.examples","command":{"q":{"_search":"38334d64ad93602d"},"limit":0},"planSummary":"EOF","keysExamined":0,"docsExamined":0,"ndeleted":0,"numYields":0,"locks":{"ParallelBatchWriterMode":{"acquireCount":{"r":1}},"ReplicationStateTransition":{"acquireCount":{"w":1}},"Global":{"acquireCount":{"w":1}},"Database":{"acquireCount":{"w":1}},"Collection":{"acquireCount":{"r":1,"w":1},"acquireWaitCount":{"w":1},"timeAcquiringMicros":{"w":148447}},"Mutex":{"acquireCount":{"r":1}}},"flowControl":{"acquireCount":1,"timeAcquiringMicros":1},"durationMillis":148}}`
	loginfo := NewLogInfo("utest-xxxxxx")
	loginfo.SetVerbose(true)
	if stat, err := loginfo.ParseLogv2(str); err != nil {
		t.Fatal(err)
	} else if stat.filter != `{"_search":1}` {
		t.Log(stat.filter)
		t.Fatal(stat.filter)
	}
}

func TestParseLogv2BSON(t *testing.T) {
	var err error
	var doc struct {
		Attributes struct {
			Command            bson.M `bson:"command"`
			Milli              int    `bson:"durationMillis"`
			NS                 string `bson:"ns"`
			OriginatingCommand bson.M `bson:"originatingCommand"`
			PlanSummary        string `bson:"planSummary"`
			Type               string `bson:"type"`
		} `bson:"attr,omitempty"`
		Component string    `bson:"c"`
		ID        int       `bson:"id"`
		Message   string    `bson:"msg"`
		Severity  string    `bson:"s"`
		Timestamp time.Time `bson:"t"`
	}
	var str = `{"t":{"$date":"2020-06-25T15:15:40.237+00:00"},"s":"I","c":"COMMAND","id":51803,
	"ctx":"conn23242","msg":"Slow query","attr":{"type":"command","ns":"Reconciliation.reconciliationRunObject",
	"appName":"MongoDB Automation Agent v10.14.8.6391 (git: 93553510ac93fbc5eeb2f7052e9b56e2b330f5b1)",
	"command":{"getMore":8443825196097945957,"collection":"reconciliationRunObject",
	"lsid":{"id":{"$uuid":"4397f6ad-0e47-4555-a5cf-d2b58c5ecb85"},
	"uid":{"$binary":{"base64":"tSN0bhb+GiHRch7NNrMh+4sLBvhgr0mdCVyGZq7FjG4=","subType":"0"}}},
	"$clusterTime":{"clusterTime":{"$timestamp":{"t":1593098139,"i":7}},
	"signature":{"hash":{"$binary":{"base64":"bgJX8H3DD0bW2uJCJgoRTKCEbMQ=","subType":"0"}},
	"keyId":6841613758402396181}},"$audit":{"$impersonatedUsers":[{"user":"mms-automation","db":"admin"}],
	"$impersonatedRoles":[{"role":"readWriteAnyDatabase","db":"admin"},{"role":"restore","db":"admin"},
	{"role":"dbAdminAnyDatabase","db":"admin"},{"role":"userAdminAnyDatabase","db":"admin"},
	{"role":"clusterAdmin","db":"admin"},{"role":"backup","db":"admin"}]},
	"$client":{"driver":{"name":"mongo-go-driver","version":"v1.1.4"},"os":{"type":"linux","architecture":"amd64"},
	"platform":"go1.14.2","application":{"name":"MongoDB Automation Agent v10.14.8.6391 (git: 93553510ac93fbc5eeb2f7052e9b56e2b330f5b1)"},
	"mongos":{"host":"atlas-pzn1iz-shard-00-00.dgipl.mongodb.net:27016","client":"192.168.254.234:43640","version":"4.4.0-rc9"}},
	"$configServerState":{"opTime":{"ts":{"$timestamp":{"t":1593098139,"i":1}},"t":2}},"$db":"Reconciliation"},"originatingCommand":{"aggregate":"reconciliationRunObject","pipeline":[{"$sample":{"size":1000}},{"$limit":1000}],"fromMongos":true,"needsMerge":true,"collation":{"locale":"simple"},"cursor":{"batchSize":0},"maxTimeMS":15000,"runtimeConstants":{"localNow":{"$date":"2020-06-25T15:15:39.932Z"},"clusterTime":{"$timestamp":{"t":1593098139,"i":7}}},"use44SortKeys":true,"useNewUpsert":true,"readConcern":{"provenance":"implicitDefault"},"writeConcern":{"w":1,"wtimeout":0,"provenance":"implicitDefault"},"shardVersion":[{"$timestamp":{"t":606,"i":0}},{"$oid":"5ef2624f583c89157d183ebf"}],"clientOperationKey":{"$uuid":"5270f3a7-2523-4cd3-9bb5-c5c8c66d4106"},"lsid":{"id":{"$uuid":"4397f6ad-0e47-4555-a5cf-d2b58c5ecb85"},"uid":{"$binary":{"base64":"tSN0bhb+GiHRch7NNrMh+4sLBvhgr0mdCVyGZq7FjG4=","subType":"0"}}},"$clusterTime":{"clusterTime":{"$timestamp":{"t":1593098139,"i":7}},"signature":{"hash":{"$binary":{"base64":"bgJX8H3DD0bW2uJCJgoRTKCEbMQ=","subType":"0"}},"keyId":6841613758402396181}},"$audit":{"$impersonatedUsers":[{"user":"mms-automation","db":"admin"}],"$impersonatedRoles":[{"role":"readWriteAnyDatabase","db":"admin"},{"role":"restore","db":"admin"},{"role":"dbAdminAnyDatabase","db":"admin"},{"role":"userAdminAnyDatabase","db":"admin"},{"role":"clusterAdmin","db":"admin"},{"role":"backup","db":"admin"}]},"$client":{"driver":{"name":"mongo-go-driver","version":"v1.1.4"},"os":{"type":"linux","architecture":"amd64"},"platform":"go1.14.2","application":{"name":"MongoDB Automation Agent v10.14.8.6391 (git: 93553510ac93fbc5eeb2f7052e9b56e2b330f5b1)"},"mongos":{"host":"atlas-pzn1iz-shard-00-00.dgipl.mongodb.net:27016","client":"192.168.254.234:43640","version":"4.4.0-rc9"}},"$configServerState":{"opTime":{"ts":{"$timestamp":{"t":1593098139,"i":1}},"t":2}},"$db":"Reconciliation"},"planSummary":"QUEUED_DATA, MULTI_ITERATOR","cursorid":8443825196097945957,"keysExamined":0,"docsExamined":0,"cursorExhausted":true,"numYields":6,"nreturned":1000,"reslen":1596235,"locks":{"ReplicationStateTransition":{"acquireCount":{"w":8}},"Global":{"acquireCount":{"r":8}},"Database":{"acquireCount":{"r":8}},"Collection":{"acquireCount":{"r":7}},"Mutex":{"acquireCount":{"r":2}}},"readConcern":{"provenance":"implicitDefault"},"writeConcern":{"w":1,"wtimeout":0,"provenance":"implicitDefault"},"storage":{"data":{"bytesRead":33429400,"timeReadingMicros":38727}},"protocol":"op_msg","durationMillis":114}}`
	if err = bson.UnmarshalExtJSON([]byte(str), false, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Component != "COMMAND" {
		t.Fatal(`expected COMMAND but got`, doc.Component)
	}
}
