// Copyright 2018 Kuei-chun Chen. All rights reserved.

package ftdc

import (
	"io/ioutil"
	"testing"
)

func TestDecode(t *testing.T) {
	var err error
	var buffer []byte

	filename := "../test_data/diagnostic.data/metrics.2017-10-12T20-08-53Z-00000"
	if buffer, err = ioutil.ReadFile(filename); err != nil {
		t.Fatal(err)
	}
	m := NewMetrics()
	m.ReadAllMetrics(buffer)
	if len(m.Blocks) == 0 {
		t.Fatal()
	}

	if _, err = m.decode([]byte(m.Blocks[0])); err != nil {
		t.Fatal(err)
	}
}
