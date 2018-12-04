// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"encoding/json"
	"testing"
)

func TestReadDiagnosticDir(t *testing.T) {
	var err error
	d := NewDiagnosticData(300)
	if err = d.readDiagnosticDir("../test_data/diagnostic.data/"); err != nil {
		t.Fatal(err)
	}
	b, _ := json.MarshalIndent(d.ServerInfo, "", "  ")
	t.Log(string(b))
	t.Log("serverStatus length", len(d.ServerStatusList))
	t.Log("replSetStatus length", len(d.ReplSetStatusList))
	span := int(d.ServerStatusList[(len(d.ServerStatusList)-1)].LocalTime.Sub(d.ServerStatusList[0].LocalTime).Seconds()) / 20
	if PrintAllStats(d.ServerStatusList, span) == "" { // every 10 minutes
		t.Fatal()
	}
}

func TestReadDiagnosticFile(t *testing.T) {
	var err error
	var diag DiagnosticData
	d := NewDiagnosticData(300)
	if diag, err = d.readDiagnosticFile("../test_data/diagnostic.data/metrics.2017-10-12T20-08-53Z-00000"); err != nil {
		t.Fatal(err)
	}
	b, _ := json.MarshalIndent(diag.ServerInfo, "", "  ")
	t.Log(string(b))
	t.Log("serverStatus length", len(diag.ServerStatusList))
	t.Log("replSetStatus length", len(diag.ReplSetStatusList))
	span := int(diag.ServerStatusList[(len(diag.ServerStatusList)-1)].LocalTime.Sub(diag.ServerStatusList[0].LocalTime).Seconds()) / 20
	if PrintAllStats(diag.ServerStatusList, span) == "" { // every 10 minutes
		t.Fatal()
	}
}

func TestAnalyzeServerStatus(t *testing.T) {
	var filename = "../test_data/tmp/keyhole_stats.2018-12-04T080240-standalone.gz"
	var err error

	d := NewDiagnosticData(300)
	if err = d.analyzeServerStatus(filename); err != nil {
		t.Fatal(err)
	}
	t.Log(len(d.ServerStatusList))
}
