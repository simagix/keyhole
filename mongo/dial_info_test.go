// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mongo

import (
	"os"
	"testing"

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
	if session, err = mgo.DialWithInfo(dialInfo); err != nil {
		return session, err
	}
	return session, err
}

func TestParseURL(t *testing.T) {
	var err error
	var dialInfo *mgo.DialInfo

	url := "mongodb+srv://user:secret@keyhole-jgtm2.mongodb.net/keyhole"
	if dialInfo, err = ParseURL(url); err != nil {
		t.Fatal(err)
	}
	if dialInfo.Username != "user" {
		t.Fatal()
	}

	url = "mongodb://user:secret@keyhole-jgtm2.mongodb.net/keyhole"
	if dialInfo, err = ParseURL(url); err != nil {
		t.Fatal(err)
	}
	if dialInfo.Username != "user" {
		t.Fatal()
	}
}

func TestAddCertificates(t *testing.T) {
}
