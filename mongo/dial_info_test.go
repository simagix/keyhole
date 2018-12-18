// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mongo

import (
	"testing"

	"github.com/globalsign/mgo"
)

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
