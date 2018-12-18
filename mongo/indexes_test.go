// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mongo

import (
	"fmt"
	"testing"

	"github.com/globalsign/mgo"
)

var dbName = "keyhole"

func indexesTestInit(t *testing.T) *mgo.Session {
	var err error
	var session *mgo.Session
	if session, err = GetTestSession(); err != nil {
		return nil
	}
	c := session.DB(dbName).C("examples")
	c.EnsureIndexKey("a")
	c.EnsureIndexKey("a", "b", "c")
	c.EnsureIndexKey("a", "b")
	c.EnsureIndexKey("a", "c")
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
