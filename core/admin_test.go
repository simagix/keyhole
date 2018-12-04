// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/globalsign/mgo"
)

var unitTestURI = "mongodb://localhost/"

func getSession() (*mgo.Session, error) {
	var err error
	var dialInfo *mgo.DialInfo
	var session *mgo.Session

	if os.Getenv("DATABASE_URL") != "" {
		unitTestURI = os.Getenv("DATABASE_URL")
	}

	if dialInfo, err = ParseDialInfo(unitTestURI); err != nil {
		return session, err
	}
	return GetSession(dialInfo, false, false, "", "")
}

func TestAdminCommand(t *testing.T) {
	var err error
	var session *mgo.Session
	if session, err = getSession(); err != nil {
		t.Fatal(err)
	}

	m, er := AdminCommand(session, "dbStats")

	if er != nil {
		t.Fatal(er)
	}
	bytes, _ := json.MarshalIndent(m, "", "  ")
	t.Log(string(bytes))
}

func TestAdminCommandOnDB(t *testing.T) {
	var err error
	var session *mgo.Session
	if session, err = getSession(); err != nil {
		t.Fatal(err)
	}

	m, er := AdminCommandOnDB(session, "dbStats", "test")

	if er != nil {
		t.Fatal(er)
	}
	bytes, _ := json.MarshalIndent(m, "", "  ")
	t.Log(string(bytes))
}

func TestCleanup(t *testing.T) {
	var err error
	var session *mgo.Session
	if session, err = getSession(); err != nil {
		t.Fatal(err)
	}
	if err = Cleanup(session); err != nil {
		t.Fatal(err)
	}
}
