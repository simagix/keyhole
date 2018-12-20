// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mongo

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/globalsign/mgo"
)

func TestGetServerInfo(t *testing.T) {
	var err error
	var dialInfo *mgo.DialInfo
	var session *mgo.Session
	var info ServerInfo

	uri := "mongodb://localhost/"
	if os.Getenv("DATABASE_URL") != "" {
		uri = os.Getenv("DATABASE_URL")
	}

	if dialInfo, err = ParseURL(uri); err != nil {
		t.Fatal(err)
	}
	if session, err = mgo.DialWithInfo(dialInfo); err != nil {
		t.Fatal(err)
	}

	if info, err = GetServerInfo(session); err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	bytes, _ := json.MarshalIndent(info, "", "  ")
	t.Log(string(bytes))
}
