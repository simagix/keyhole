// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"regexp"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func TestParseLogv2(t *testing.T) {
	str := `{"t":{"$date":"2020-03-30T09:18:46.696-04:00"},"s":"I", "c":"-","id":51803,"ctx":"conn487","msg":"slow query","attr":{"type":"remove","ns":"_KEYHOLE_88800.examples","command":{"q":{"_search":"38334d64ad93602d"},"limit":0},"planSummary":"EOF","keysExamined":0,"docsExamined":0,"ndeleted":0,"numYields":0,"locks":{"ParallelBatchWriterMode":{"acquireCount":{"r":1}},"ReplicationStateTransition":{"acquireCount":{"w":1}},"Global":{"acquireCount":{"w":1}},"Database":{"acquireCount":{"w":1}},"Collection":{"acquireCount":{"r":1,"w":1},"acquireWaitCount":{"w":1},"timeAcquiringMicros":{"w":148447}},"Mutex":{"acquireCount":{"r":1}}},"flowControl":{"acquireCount":1,"timeAcquiringMicros":1},"durationMillis":148}}`
	str = `{"t":{"$date":"2020-09-28T11:13:09.234+00:00"},"s":"I",  "c":"COMMAND",  "id":51803,   "ctx":"conn4479","msg":"Slow query","attr":{"type":"command","ns":"stringr.devices","command":{"findAndModify":"devices","query":{"os":"iOS","token":"344E84F746E755C9C5544D0EA2F348BD78BD0FE9C23C802AB3B7DEEEAB4EC027","user":{"$oid":"59154269cfe1f2d40943d7f5"}},"new":false,"remove":false,"upsert":true,"fields":{},"update":{"$setOnInsert":{"__v":0},"$set":{"token":"344E84F746E755C9C5544D0EA2F348BD78BD0FE9C23C802AB3B7DEEEAB4EC027","os":"iOS","dma":"Phoenix, AZ","zipCode":"85206","city":"Mesa","stateCode":"AZ","countryCode":"US","user":{"$oid":"59154269cfe1f2d40943d7f5"},"date":{"$date":"2020-09-28T11:13:09.077Z"},"location":[-111.74620925100147,33.3791438189752],"client":"StringrIos/2.6.4","applicationId":"com.stringr.app","active":true}},"writeConcern":{"w":"majority"},"lsid":{"id":{"$uuid":"fed35edb-a198-41d9-89f2-133141aa2e25"}},"txnNumber":645,"$clusterTime":{"clusterTime":{"$timestamp":{"t":1601291589,"i":3}},"signature":{"hash":{"$binary":{"base64":"s5assGz66sijGO61QiOA+0tJC3Q=","subType":"0"}},"keyId":6877257850905690113}},"$db":"stringr"},"planSummary":"IXSCAN { token: 1, os: 1 }","keysExamined":0,"docsExamined":0,"nMatched":0,"nModified":0,"upsert":true,"keysInserted":19,"numYields":0,"queryHash":"16FE1304","planCacheKey":"ADE4584D","reslen":239,"locks":{"ParallelBatchWriterMode":{"acquireCount":{"r":2}},"ReplicationStateTransition":{"acquireCount":{"w":3}},"Global":{"acquireCount":{"w":2}},"Database":{"acquireCount":{"w":2}},"Collection":{"acquireCount":{"w":2}},"Mutex":{"acquireCount":{"r":2}}},"flowControl":{"acquireCount":1,"timeAcquiringMicros":1},"writeConcern":{"w":"majority","wtimeout":0,"provenance":"clientSupplied"},"storage":{"data":{"bytesRead":24428,"timeReadingMicros":23}},"protocol":"op_msg","durationMillis":151}}`
	loginfo := NewLogInfo("utest-xxxxxx")
	loginfo.SetVerbose(true)
	if stat, err := loginfo.ParseLogv2(str); err != nil {
		t.Fatal(err)
	} else if stat.filter != `{"os":1,"token":1,"user":1}` {
		t.Log(stat.filter)
		t.Fatal(stat.filter)
	}
}

func TestInRepeatedDocPattern(t *testing.T) {
	str := `{"a":[{"$binary":{"base64":1,"subType":1}},{"$binary":{"base64":1,"subType":1}}}],"b":[1,1,1],"c":[1]}`
	re := regexp.MustCompile(`\[1(,1)*\]`)
	str = re.ReplaceAllString(str, `[...]`)
	re = regexp.MustCompile(`\[\{\S+\}(,\{\S+\})*\]`)
	str = re.ReplaceAllString(str, `[...]`)
	if str != `{"a":[...],"b":[...],"c":[...]}` {
		t.Fatal(str)
	}
	t.Log(str)
	str = `{"t":{"$date":"2020-11-05T09:30:50.680+00:00"},"s":"I",  "c":"WRITE",    "id":51803,   "ctx":"conn851","msg":"Slow query","attr":{"type":"update","ns":"testcase.samplecollection","command":{"q":{"_id":{"$in":[{"$binary":{"base64":"jtWtXe56xEGlJ+2cNiz+Yg==","subType":"3"}},{"$binary":{"base64":"tgP+I+Jl+I1wya7jxapMAg==","subType":"3"}},{"$binary":{"base64":"ElxhlRkJfT8FOCxfe+OJ9w==","subType":"3"}},{"$binary":{"base64":"nFTaxhmUzWC/yEnrqYN3NA==","subType":"3"}}]}},"u":{"$set":{"Touched":{"$date":"2020-11-05T09:30:47.707Z"}}},"multi":true,"upsert":false},"planSummary":"IXSCAN { _id: 1 }","keysExamined":4,"docsExamined":4,"nMatched":4,"nModified":4,"numYields":4,"queryHash":"EDADD1D2","planCacheKey":"E5EEC495","locks":{"ParallelBatchWriterMode":{"acquireCount":{"r":5}},"ReplicationStateTransition":{"acquireCount":{"w":5}},"Global":{"acquireCount":{"w":5}},"Database":{"acquireCount":{"w":5}},"Collection":{"acquireCount":{"w":5}},"Mutex":{"acquireCount":{"r":1}}},"flowControl":{"acquireCount":5,"timeAcquiringMicros":49},"storage":{"data":{"bytesRead":40}},"durationMillis":234}}`
	loginfo := NewLogInfo("utest-xxxxxx")
	loginfo.SetVerbose(true)
	if stat, err := loginfo.ParseLogv2(str); err != nil {
		t.Fatal(err)
	} else if stat.filter != `{"_id":{"$in":[...]}}` {
		t.Fatal(stat.filter)
	} else {
		t.Log(stat.filter)
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
