// Copyright 2018 Kuei-chun Chen. All rights reserved.

package ftdc

import (
	"bytes"
	"testing"
)

func TestGetUint32(t *testing.T) {
	var ui uint32
	buffer := []byte{0x0f, 0x00, 0x00, 0x00}
	r := bytes.NewReader(buffer)
	ui = GetUint32(r)
	if ui != 15 {
		t.Log(ui)
		t.Fatal(ui, 15)
	}
}

func TestUvarint(t *testing.T) {
	var ui uint64
	buffer := []byte{0x0f}
	r := bytes.NewReader(buffer)
	ui = Uvarint(r)
	if ui != 15 {
		t.Log(ui)
		t.Fatal(ui, 15)
	}
}
