// Copyright 2018 Kuei-chun Chen. All rights reserved.

package ftdc

import (
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

var filename = "testdata/diagnostic.data/metrics.2017-10-12T20-08-53Z-00000"

func TestDecode(t *testing.T) {
	var err error
	var buffer []byte

	if buffer, err = os.ReadFile(filename); err != nil {
		t.Fatal(err)
	}
	m := NewMetrics()
	m.ReadAllMetrics(&buffer)
	if len(m.Data) == 0 {
		t.Fatal()
	}

	if _, err = m.decode([]byte(m.Data[0].Block)); err != nil {
		t.Fatal(err)
	}
}

func TestTraverseDocElem(t *testing.T) {
	var err error
	var buffer []byte

	if buffer, err = os.ReadFile(filename); err != nil {
		t.Fatal(err)
	}
	m := NewMetrics()
	m.ReadAllMetrics(&buffer)
	if len(m.Data) == 0 {
		t.Fatal()
	}

	var dp = MetricsData{DataPointsMap: map[string][]uint64{}}
	var docElem = bson.D{}
	var attribsList = []string{}
	bson.Unmarshal(m.Data[0].Block, &docElem) // first document
	traverseDocElem(&attribsList, &dp.DataPointsMap, docElem, "")
	if len(attribsList) == 0 || len(dp.DataPointsMap) != len(attribsList) {
		t.Fatal()
	}
}
