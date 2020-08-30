// Copyright 2019 Kuei-chun Chen. All rights reserved.

package analytics

import (
	"testing"
	"time"
)

func TestGetDataPoint(t *testing.T) {
	tm := float64(time.Now().UnixNano() / 1000 / 1000)
	v := 123.45
	dp := getDataPoint(v, tm)
	if dp[0] != v {
		t.Fatal()
	}
	t.Log(dp)
}

func TestGetServerStatusTimeSeriesDoc(t *testing.T) {
	d := NewDiagnosticData()
	var filenames = []string{DiagnosticDataFilename}
	d.DecodeDiagnosticData(filenames)
	tsd := getServerStatusTimeSeriesDoc(d.ServerStatusList)
	if len(tsd) == 0 {
		t.Fatal()
	}
}

func TestGetSystemMetricsTimeSeriesDoca(t *testing.T) {
	d := NewDiagnosticData()
	var filenames = []string{DiagnosticDataFilename}
	d.DecodeDiagnosticData(filenames)
	tsd, _ := getSystemMetricsTimeSeriesDoc(d.SystemMetricsList)
	if len(tsd) == 0 {
		t.Fatal()
	}
}

func TestGetReplSetGetStatusTimeSeriesDoc(t *testing.T) {
	d := NewDiagnosticData()
	var filenames = []string{DiagnosticDataFilename}
	d.DecodeDiagnosticData(filenames)
	tsd, _ := getReplSetGetStatusTimeSeriesDoc(d.ReplSetStatusList, &[]string{})
	if len(tsd) == 0 {
		t.Fatal()
	}
}

func TestGetWiredTigerTimeSeriesDoc(t *testing.T) {
	d := NewDiagnosticData()
	var filenames = []string{DiagnosticDataFilename}
	d.DecodeDiagnosticData(filenames)
	tsd := getWiredTigerTimeSeriesDoc(d.ServerStatusList)
	if len(tsd) == 0 {
		t.Fatal()
	}
}

func TestAddFTDCDetailStats(t *testing.T) {
	filenames := []string{"testdata/diagnostic.data/metrics.2018-10-07T15-15-22Z-00000"}
	metrics := NewMetrics()
	diag := NewDiagnosticData()
	diag.DecodeDiagnosticData(filenames)
	metrics.AddFTDCDetailStats(diag)
}
