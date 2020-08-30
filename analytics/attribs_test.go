// Copyright 2020 Kuei-chun Chen. All rights reserved.

package analytics

import (
	"io/ioutil"
	"testing"

	"github.com/simagix/keyhole/ftdc"
)

func TestGetServerStatusDataPoints(t *testing.T) {
	var err error
	var buffer []byte

	if buffer, err = ioutil.ReadFile(DiagnosticDataFilename); err != nil {
		t.Fatal(err)
	}
	metrics := ftdc.NewMetrics()
	metrics.ReadAllMetrics(&buffer)
	attrib := NewAttribs(&metrics.Data[0].DataPointsMap)
	v := attrib.GetServerStatusDataPoints(0)
	t.Log(v)
}

func TestGetSystemMetricsDataPoints(t *testing.T) {
	var err error
	var buffer []byte

	if buffer, err = ioutil.ReadFile(DiagnosticDataFilename); err != nil {
		t.Fatal(err)
	}
	metrics := ftdc.NewMetrics()
	metrics.ReadAllMetrics(&buffer)
	attrib := NewAttribs(&metrics.Data[0].DataPointsMap)
	v := attrib.GetSystemMetricsDataPoints(0)
	t.Log(v)
}
