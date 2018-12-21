// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mongo

import (
	"encoding/json"
	"testing"

	"github.com/globalsign/mgo"
	"github.com/mongodb/mongo-go-driver/bson"
)

func TestRunAdminCommand(t *testing.T) {
	var err error
	var session *mgo.Session
	var m bson.M
	if session, err = GetTestSession(); err != nil {
		t.Fatal(err)
	}

	if m, err = RunAdminCommand(session, "dbStats"); err != nil {
		t.Fatal(err)
	} else {
		bytes, _ := json.MarshalIndent(m, "", "  ")
		t.Log(string(bytes))
	}
}

func TestRunCommandOnDB(t *testing.T) {
	var err error
	var session *mgo.Session
	var m bson.M
	if session, err = GetTestSession(); err != nil {
		t.Fatal(err)
	}

	if m, err = RunCommandOnDB(session, "dbStats", "test"); err != nil {
		t.Fatal(err)
	} else {
		bytes, _ := json.MarshalIndent(m, "", "  ")
		t.Log(string(bytes))
	}
}

func TestIsMaster(t *testing.T) {
	var err error
	var session *mgo.Session
	var m bson.M
	if session, err = GetTestSession(); err != nil {
		t.Fatal(err)
	}
	if m, err = IsMaster(session); err != nil {
		t.Fatal(err)
	}
	t.Log(m)
}
