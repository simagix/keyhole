// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/globalsign/mgo"
	"github.com/simagix/keyhole/mongo"
)

func TestGetSession(t *testing.T) {
	var err error
	var dialInfo *mgo.DialInfo
	var session *mgo.Session

	uri := "mongodb://localhost/"
	if os.Getenv("DATABASE_URL") != "" {
		uri = os.Getenv("DATABASE_URL")
	}

	if dialInfo, err = mongo.ParseURL(uri); err != nil {
		t.Fatal(err)
	}
	if session, err = mongo.GetSession(dialInfo, false, false, "", ""); err != nil {
		t.Fatal(err)
	}

	t.Log(session)
}

func TestGetSessionTLS(t *testing.T) {
	var err error
	var dialInfo *mgo.DialInfo
	var session *mgo.Session

	uri := "mongodb://localhost/"
	if os.Getenv("DATABASE_URL") != "" {
		uri = os.Getenv("DATABASE_URL")
	}

	if dialInfo, err = mongo.ParseURL(uri); err != nil {
		t.Fatal(err)
	}
	if session, err = mongo.GetSession(dialInfo, false, false, "/etc/ssl/certs/ca.pem", "/etc/ssl/certs/server.pem"); err != nil {
		t.Fatal(err)
	}

	t.Log(session)
}

func TestGetMongoServerInfo(t *testing.T) {
	var err error
	var dialInfo *mgo.DialInfo
	var session *mgo.Session
	var info MongoServerInfo

	uri := "mongodb://localhost/"
	if os.Getenv("DATABASE_URL") != "" {
		uri = os.Getenv("DATABASE_URL")
	}

	if dialInfo, err = mongo.ParseURL(uri); err != nil {
		t.Fatal(err)
	}
	if session, err = mongo.GetSession(dialInfo, false, false, "", ""); err != nil {
		t.Fatal(err)
	}

	if info, err = GetMongoServerInfo(session); err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	bytes, _ := json.MarshalIndent(info, "", "  ")
	t.Log(string(bytes))
}
