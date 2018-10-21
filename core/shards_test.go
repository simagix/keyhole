// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"testing"

	"github.com/globalsign/mgo"
)

func TestGetShards(t *testing.T) {
	var err error
	var list []string
	var session *mgo.Session
	if session, err = getSession(); err != nil {
		t.Fatal(err)
	}

	if list, err = GetShards(session, unitTestURI); err != nil {
		t.Fatal(err)
	}

	t.Log(list)
}

func TestShardCollection(t *testing.T) {
	var err error
	var session *mgo.Session
	if session, err = getSession(); err != nil {
		t.Fatal(err)
	}
	if err = ShardCollection(session); err != nil {
		t.Fatal(err)
	}
}
