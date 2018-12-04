// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"encoding/json"
	"testing"

	"github.com/globalsign/mgo"
)

func TestGetTransactions(t *testing.T) {
	TransactionDoc := GetTransactions("")
	bytes, _ := json.MarshalIndent(TransactionDoc, "", "  ")
	t.Log(string(bytes))
}

func TestExecTXForDemo(t *testing.T) {
	var err error
	var session *mgo.Session
	if session, err = getSession(); err != nil {
		t.Fatal(err)
	}
	c := session.DB(SimDBName).C(CollectionName)
	n := execTXForDemo(c, GetDemoDoc())
	if n != 5 {
		t.Fatal()
	}
}

func TestExecTXByTemplate(t *testing.T) {
	var err error
	var session *mgo.Session
	if session, err = getSession(); err != nil {
		t.Fatal(err)
	}
	c := session.DB(SimDBName).C(CollectionName)
	n := execTXByTemplate(c, GetDemoDoc())
	if n != 5 {
		t.Fatal()
	}
}
