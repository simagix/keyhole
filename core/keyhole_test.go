// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"testing"

	"github.com/globalsign/mgo"
)

func TestGetShardsURIList(t *testing.T) {
	var err error
	var session *mgo.Session
	if session, err = getSession(); err != nil {
		t.Fatal(err)
	}
	uri := "mongodb://localhost/"
	uriList, err := getShardsURIList(session, uri)
	if err != nil {
		t.Fatal()
	}

	t.Log(uriList)
}
