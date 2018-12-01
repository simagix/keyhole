// Copyright 2018 Kuei-chun Chen. All rights reserved.

package ftdc

import (
	"bytes"
	"testing"
)

func TestUvarint(t *testing.T) {
	var ui uint64
	buffer := []byte{0x0f, 0x01, 0x08, 0xff, 0x08, 0x01, 0xab, 0x01}
	r := bytes.NewReader(buffer)
	ui = Uvarint(r)
	if ui != 15 {
		t.Fatal(ui, 15)
	}
}
