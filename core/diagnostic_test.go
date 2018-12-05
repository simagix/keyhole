// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

const testDataDirectory = "../test_data/"
const diagnosticDataDirectory = testDataDirectory + "diagnostic.data/"
const diagnosticDataFilename = diagnosticDataDirectory + "metrics.2017-10-12T20-08-53Z-00000"
const keyholeStatsFilename = testDataDirectory + "tmp/keyhole_stats.2018-12-04T080240-standalone.gz"

func TestReadDiagnosticDir(t *testing.T) {
	var err error
	d := NewDiagnosticData(300)
	if err = d.readDiagnosticDir(diagnosticDataDirectory); err != nil {
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

func TestReadDiagnosticFiles(t *testing.T) {
	var err error
	var files []os.FileInfo
	var filenames []string

	if files, err = ioutil.ReadDir(diagnosticDataDirectory); err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		if strings.Index(f.Name(), "metrics.") != 0 && strings.Index(f.Name(), "keyhole_stats.") != 0 {
			continue
		}
		filename := diagnosticDataDirectory + "/" + f.Name()
		filenames = append(filenames, filename)
	}
	d := NewDiagnosticData(300)
	if err = d.readDiagnosticFiles(filenames); err != nil {
		t.Fatal(err)
	}
}

func TestReadDiagnosticFile(t *testing.T) {
	var err error
	var diag DiagnosticData
	d := NewDiagnosticData(300)
	if diag, err = d.readDiagnosticFile(diagnosticDataFilename); err != nil {
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
	var err error

	d := NewDiagnosticData(300)
	if err = d.analyzeServerStatus(keyholeStatsFilename); err != nil {
		t.Fatal(err)
	}
	t.Log(len(d.ServerStatusList))
}
