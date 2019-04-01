// Copyright 2018 Kuei-chun Chen. All rights reserved.

package web

import (
	"testing"
	"time"

	"github.com/simagix/keyhole/sim"
)

const TestDataDirectory = "testdata/"
const DiagnosticDataDirectory = TestDataDirectory + "diagnostic.data/"
const DiagnosticDataFilename = DiagnosticDataDirectory + "metrics.2017-10-19T20-03-15Z-00000"

func TestGetDataPoint(t *testing.T) {
	tm := float64(time.Now().UnixNano() / 1000 / 1000)
	v := 123.45
	dp := getDataPoint(v, tm)
	if dp[0] != v {
		t.Fatal()
	}
	t.Log(dp)
}

func TestSetFTDCSummaryStats(t *testing.T) {
	d := sim.NewDiagnosticData(300)
	var filenames = []string{DiagnosticDataFilename}
	d.DecodeDiagnosticData(filenames)
	g.SetFTDCSummaryStats(d)
}

func TestInitServerStatusTimeSeriesDoc(t *testing.T) {
	d := sim.NewDiagnosticData(300)
	var filenames = []string{DiagnosticDataFilename}
	d.DecodeDiagnosticData(filenames)
	tsd := initServerStatusTimeSeriesDoc(d.ServerStatusList)
	if len(tsd) == 0 {
		t.Fatal()
	}
}

func TestInitSystemMetricsTimeSeriesDoca(t *testing.T) {
	d := sim.NewDiagnosticData(300)
	var filenames = []string{DiagnosticDataFilename}
	d.DecodeDiagnosticData(filenames)
	tsd, _ := initSystemMetricsTimeSeriesDoc(d.SystemMetricsList)
	if len(tsd) == 0 {
		t.Fatal()
	}
}

func TestInitReplSetGetStatusTimeSeriesDoc(t *testing.T) {
	d := sim.NewDiagnosticData(300)
	var filenames = []string{DiagnosticDataFilename}
	d.DecodeDiagnosticData(filenames)
	tsd, _ := initReplSetGetStatusTimeSeriesDoc(d.ReplSetStatusList)
	if len(tsd) == 0 {
		t.Fatal()
	}
}

func TestInitWiredTigerTimeSeriesDoc(t *testing.T) {
	d := sim.NewDiagnosticData(300)
	var filenames = []string{DiagnosticDataFilename}
	d.DecodeDiagnosticData(filenames)
	tsd := initWiredTigerTimeSeriesDoc(d.ServerStatusList)
	if len(tsd) == 0 {
		t.Fatal()
	}
}
