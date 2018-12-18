// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mongo

import (
	"fmt"
	"os"
	"testing"

	"github.com/globalsign/mgo"
	keyhole "github.com/simagix/keyhole/core"
)

var dbName = "keyhole"

func indexesTestInit(t *testing.T) *mgo.Session {
	var err error
	var dialInfo *mgo.DialInfo
	var session *mgo.Session
	uri := "mongodb://localhost/"
	if os.Getenv("DATABASE_URL") != "" {
		uri = os.Getenv("DATABASE_URL")
	}

	if dialInfo, err = keyhole.ParseDialInfo(uri); err != nil {
		t.Fatal(err)
	}
	if session, err = keyhole.GetSession(dialInfo, false, false, "", ""); err != nil {
		t.Fatal(err)
	}

	c := session.DB(dbName).C("vehicles")
	c.EnsureIndexKey("a")
	c.EnsureIndexKey("a", "b")
	c.EnsureIndexKey("a", "-b")
	return session
}

func TestGetIndexesFromDB(t *testing.T) {
	session := indexesTestInit(t)
	defer session.Close()
	var str string

	// get index from keyhole database
	str = GetIndexesFromDB(session, dbName)
	fmt.Println(str)
	t.Log(str)
}

func TestGetIndexes(t *testing.T) {
	session := indexesTestInit(t)
	defer session.Close()
	var str string
	// get all indexes
	str = GetIndexes(session, "")
	t.Log(str)
}
