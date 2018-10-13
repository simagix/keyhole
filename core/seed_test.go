// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"errors"
	"os"
	"testing"

	"github.com/globalsign/mgo"
)

func TestSeed(t *testing.T) {
	var err error
	var dialInfo *mgo.DialInfo
	var session *mgo.Session
	uri := "mongodb://localhost/"
	if os.Getenv("DATABASE_URL") != "" {
		uri = os.Getenv("DATABASE_URL")
	}

	if dialInfo, err = mgo.ParseURL(uri); err != nil {
		t.Fatal(err)
	}
	if session, err = GetSession(dialInfo, false, "", ""); err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	file := ""
	collection := ""
	total := 1000
	isDrop := true
	dbName := "keyhole"
	verbose := false

	sb := NewSeedBase(file, collection, total, isDrop, dbName, verbose)

	// calling seed()
	sb.SeedData(session)
	db := session.DB(dbName)
	coll := db.C("cars")
	var count int
	if count, err = coll.Count(); err != nil {
		t.Fatal(err)
	}

	if total != count {
		t.Fatal(errors.New("cars count doesn't match"))
	}
}

func TestSeedFromTemplate(t *testing.T) {
	var err error
	var dialInfo *mgo.DialInfo
	var session *mgo.Session
	uri := "mongodb://localhost/"
	if os.Getenv("DATABASE_URL") != "" {
		uri = os.Getenv("DATABASE_URL")
	}

	if dialInfo, err = mgo.ParseURL(uri); err != nil {
		t.Fatal(err)
	}
	if session, err = GetSession(dialInfo, false, "", ""); err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	file := "../examples/template.json"
	collection := "template"
	total := 1000
	isDrop := true
	dbName := "keyhole"
	verbose := false

	sb := NewSeedBase(file, collection, total, isDrop, dbName, verbose)

	// calling seedFromTemplate
	sb.SeedData(session)
	db := session.DB(dbName)
	coll := db.C("cars")
	var count int
	if count, err = coll.Count(); err != nil {
		t.Fatal(err)
	}

	if total != count {
		t.Fatal(errors.New("cars count doesn't match"))
	}
}
