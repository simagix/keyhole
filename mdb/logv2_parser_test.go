// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"testing"
)

func TestParseLogv2(t *testing.T) {
	str := `{"t":{"$date":"2020-03-30T09:18:46.696-04:00"},"s":"I", "c":"-",       "id":51803,"ctx":"conn487","msg":"slow query","attr":{"type":"remove","ns":"_KEYHOLE_88800.examples","command":{"q":{"_search":"38334d64ad93602d"},"limit":0},"planSummary":"EOF","keysExamined":0,"docsExamined":0,"ndeleted":0,"numYields":0,"locks":{"ParallelBatchWriterMode":{"acquireCount":{"r":1}},"ReplicationStateTransition":{"acquireCount":{"w":1}},"Global":{"acquireCount":{"w":1}},"Database":{"acquireCount":{"w":1}},"Collection":{"acquireCount":{"r":1,"w":1},"acquireWaitCount":{"w":1},"timeAcquiringMicros":{"w":148447}},"Mutex":{"acquireCount":{"r":1}}},"flowControl":{"acquireCount":1,"timeAcquiringMicros":1},"durationMillis":148}}`
	loginfo := NewLogInfo()
	if stat, err := loginfo.ParseLogv2(str); err != nil {
		t.Fatal(err)
	} else if stat.filter != `{"_search":1}` {
		t.Fatal(stat.filter)
	}
}
