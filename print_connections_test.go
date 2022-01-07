// Copyright 2020-present Kuei-chun Chen. All rights reserved.

package keyhole

import "testing"

func TestPrintConnectionsFromURI(t *testing.T) {
	var err error
	uri := "mongodb://user:password@localhost/"
	if err = PrintConnectionsFromURI(uri); err != nil {
		t.Fatal(err)
	}
}
