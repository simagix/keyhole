// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

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

func TestReinitGrafana(t *testing.T) {
	d := NewDiagnosticData(300)
	var filenames = []string{diagnosticDataFilename}
	d.PrintDiagnosticData(filenames, true)
	g.ReinitGrafana(d)
}

func TestInitServerStatusTimeSeriesDoc(t *testing.T) {
	d := NewDiagnosticData(300)
	var filenames = []string{diagnosticDataFilename}
	d.PrintDiagnosticData(filenames, true)
	tsd := g.initServerStatusTimeSeriesDoc(d.ServerStatusList)
	if len(tsd) == 0 {
		t.Fatal()
	}
}

func TestInitSystemMetricsTimeSeriesDoca(t *testing.T) {
	d := NewDiagnosticData(300)
	var filenames = []string{diagnosticDataFilename}
	d.PrintDiagnosticData(filenames, true)
	tsd, disks := g.initSystemMetricsTimeSeriesDoc(d.SystemMetricsList)
	if len(tsd) == 0 {
		t.Fatal()
	}
	t.Log(disks)
}

func TestInitReplSetGetStatusTimeSeriesDoc(t *testing.T) {
	d := NewDiagnosticData(300)
	var filenames = []string{diagnosticDataFilename}
	d.PrintDiagnosticData(filenames, true)
	tsd, lags := g.initReplSetGetStatusTimeSeriesDoc(d.ReplSetStatusList)
	if len(tsd) == 0 {
		t.Fatal()
	}
	t.Log(lags)
}

func TestInitWiredTigerTimeSeriesDoc(t *testing.T) {
	d := NewDiagnosticData(300)
	var filenames = []string{diagnosticDataFilename}
	d.PrintDiagnosticData(filenames, true)
	tsd := g.initWiredTigerTimeSeriesDoc(d.ServerStatusList)
	if len(tsd) == 0 {
		t.Fatal()
	}
}
