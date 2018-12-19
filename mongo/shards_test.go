// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mongo

import (
	"testing"

	"github.com/globalsign/mgo"
)

func TestGetShards(t *testing.T) {
	var err error
	var list []string
	var session *mgo.Session
	if session, err = GetTestSession(); err != nil {
		t.Fatal(err)
	}

	if list, err = GetShards(session, UnitTestURL); err != nil {
		t.Log(err)
		return
	}

	t.Log(list)
}
