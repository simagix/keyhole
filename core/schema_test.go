// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"os"
	"testing"

	"github.com/globalsign/mgo"
	"github.com/simagix/keyhole/mongo"
)

func TestGetSchemaFromCollection(t *testing.T) {
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
	defer session.Close()

	var str string
	if str, err = GetSchemaFromCollection(session, "keyhole", "cars", false); err != nil {
		t.Fatal(err)
	}

	t.Log(str)
}
