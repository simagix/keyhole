// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"os"
	"testing"

	"github.com/globalsign/mgo"
	"github.com/simagix/keyhole/mongo"
)

var unitTestURI = "mongodb://localhost/"

func getSession() (*mgo.Session, error) {
	var err error
	var dialInfo *mgo.DialInfo
	var session *mgo.Session

	if os.Getenv("DATABASE_URL") != "" {
		unitTestURI = os.Getenv("DATABASE_URL")
	}

	if dialInfo, err = mongo.ParseURL(unitTestURI); err != nil {
		return session, err
	}
	return mongo.GetSession(dialInfo, "", "")
}

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
