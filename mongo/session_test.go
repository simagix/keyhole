// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mongo

import (
	"os"
	"testing"

	"github.com/globalsign/mgo"
)

func GetTestSession() (*mgo.Session, error) {
	var err error
	var dialInfo *mgo.DialInfo
	var session *mgo.Session
	url := "mongodb://localhost/"
	if os.Getenv("DATABASE_URL") != "" {
		url = os.Getenv("DATABASE_URL")
	}

	if dialInfo, err = ParseURL(url); err != nil {
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
	if session, err = GetSessionWithTimeout(dialInfo, false, false, "", "", 30); err != nil {
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
