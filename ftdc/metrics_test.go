// Copyright 2018 Kuei-chun Chen. All rights reserved.

package ftdc

import (
	"os"
	"testing"
)

func TestReadAllMetrics(t *testing.T) {
	var err error
	var buffer []byte
	if buffer, err = os.ReadFile(filename); err != nil {
		t.Fatal(err)
	}
	m := NewMetrics()
	m.ReadAllMetrics(&buffer)
	if len(m.Data) != 164 {
		t.Fatal()
	}
}
