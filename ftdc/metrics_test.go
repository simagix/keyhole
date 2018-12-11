// Copyright 2018 Kuei-chun Chen. All rights reserved.

package ftdc

import (
	"io/ioutil"
	"testing"
)

func TestReadMetricsSummary(t *testing.T) {
	var err error
	var buffer []byte
	if buffer, err = ioutil.ReadFile(filename); err != nil {
		t.Fatal(err)
	}
	m := NewMetrics()
	m.ReadMetricsSummary(buffer)
	if len(m.Data) != 164 {
		t.Fatal()
	}
}

func TestReadAllMetrics(t *testing.T) {
	var err error
	var buffer []byte
	if buffer, err = ioutil.ReadFile(filename); err != nil {
		t.Fatal(err)
	}
	m := NewMetrics()
	m.ReadAllMetrics(buffer)
	if len(m.Data) != 164 {
		t.Fatal()
	}
}

func TestReadMetrics(t *testing.T) {
	var err error
	var buffer []byte
	if buffer, err = ioutil.ReadFile(filename); err != nil {
		t.Fatal(err)
	}
	m := NewMetrics()
	m.readMetrics(buffer, false)
	if len(m.Data) != 164 {
		t.Fatal()
	}
}
