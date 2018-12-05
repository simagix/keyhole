// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"io/ioutil"
	"testing"

	"github.com/simagix/keyhole/ftdc"
)

func TestGetServerStatusDataPoints(t *testing.T) {
	var err error
	var buffer []byte

	if buffer, err = ioutil.ReadFile(diagnosticDataFilename); err != nil {
		t.Fatal(err)
	}
	metrics := ftdc.NewMetrics()
	metrics.ReadAllMetrics(buffer)
	v := getServerStatusDataPoints(metrics.Data[0].DataPointsMap, 0)
	t.Log(v)
}

func TestGetSystemMetricsDataPoints(t *testing.T) {
	var err error
	var buffer []byte

	if buffer, err = ioutil.ReadFile(diagnosticDataFilename); err != nil {
		t.Fatal(err)
	}
	metrics := ftdc.NewMetrics()
	metrics.ReadAllMetrics(buffer)
	v := getSystemMetricsDataPoints(metrics.Data[0].DataPointsMap, 0)
	t.Log(v)
}
