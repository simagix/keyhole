// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mongo

import (
	"os"
	"testing"
	"time"

	"github.com/globalsign/mgo"
)

var UnitTestURL = "mongodb://localhost/"

func GetTestSession() (*mgo.Session, error) {
	var err error
	var dialInfo *mgo.DialInfo
	var session *mgo.Session

	if os.Getenv("DATABASE_URL") != "" {
		UnitTestURL = os.Getenv("DATABASE_URL")
	}

	if dialInfo, err = ParseURL(UnitTestURL); err != nil {
		return session, err
	}
	if session, err = GetSession(dialInfo, false, false, "", ""); err != nil {
		return session, err
	}
	return session, err
}

func TestGetSession(t *testing.T) {
	var err error
	var dialInfo *mgo.DialInfo
	var session *mgo.Session

	uri := "mongodb://localhost/"
	if os.Getenv("DATABASE_URL") != "" {
		uri = os.Getenv("DATABASE_URL")
	}

	if dialInfo, err = ParseURL(uri); err != nil {
		t.Fatal(err)
	}
	if session, err = GetSession(dialInfo, false, false, "", ""); err != nil {
		t.Fatal(err)
	}

	t.Log(session)
}

func TestGetSessionWithTimeout(t *testing.T) {
	var err error
	var dialInfo *mgo.DialInfo
	var session *mgo.Session

	uri := "mongodb://localhost/"
	if os.Getenv("DATABASE_URL") != "" {
		uri = os.Getenv("DATABASE_URL")
	}

	if dialInfo, err = ParseURL(uri); err != nil {
		t.Fatal(err)
	}

	dialInfo.Timeout = time.Duration(30 * time.Second)
	if session, err = GetSession(dialInfo, false, false, "", ""); err != nil {
		t.Log(err)
	}

	t.Log(session)
}

func TestGetSessionWithCertificates(t *testing.T) {
	var err error
	var dialInfo *mgo.DialInfo
	var session *mgo.Session

	uri := "mongodb://localhost/"
	if os.Getenv("DATABASE_URL") != "" {
		uri = os.Getenv("DATABASE_URL")
	}

	if dialInfo, err = ParseURL(uri); err != nil {
		t.Fatal(err)
	}
	if session, err = GetSession(dialInfo, false, false, "/etc/ssl/certs/ca.pem", "/etc/ssl/certs/server.pem"); err != nil {
		t.Fatal(err)
	}

	t.Log(session)
}
