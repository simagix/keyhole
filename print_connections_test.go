// Copyright 2020-present Kuei-chun Chen. All rights reserved.

package keyhole

import "testing"

func TestPrintConnectionsFromFile(t *testing.T) {
	// {
	// 	"action": "print_connections",
	// 	"filename": "testdata/mongod.log.gz"
	// }
	var err error
	filename := "testdata/mongod.log.gz"
	if err = PrintConnectionsFromFile(filename); err != nil {
		t.Fatal(err)
	}
}

func TestPrintConnectionsFromURI(t *testing.T) {
	// {
	// 	"action": "print_connections",
	// 	"uri": "mongodb+srv://user:password@cluster.xxxxxx.mongodb.net/"
	// }
	var err error
	uri := "mongodb://user:password@localhost/"
	if err = PrintConnectionsFromURI(uri); err != nil {
		t.Fatal(err)
	}
}
