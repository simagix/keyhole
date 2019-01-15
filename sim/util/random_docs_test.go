// Copyright 2018 Kuei-chun Chen. All rights reserved.

package util

import (
	"reflect"
	"testing"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

func TestGetDocByTemplate(t *testing.T) {
	var err error
	var doc bson.M

	if doc, err = GetDocByTemplate("testdata/template.json", true); err != nil {
		t.Fatal(err)
	}

	if doc["_id"] != "$oId" {
		t.Fatal("expected $oid but got", doc["_id"])
	} else if doc["lastUpdated"] != "$date" {
		t.Fatal("expected $date but got", doc["lastUpdated"])
	}

	if doc, err = GetDocByTemplate("testdata/template.json", false); err != nil {
		t.Fatal(err)
	}

	_, ok := doc["_id"].(primitive.ObjectID)
	if !ok {
		t.Fatal("expected ObjectId but got", reflect.TypeOf(doc["_id"]))
	}
}

func TestGetMagicString(t *testing.T) {
	email := getMagicString("ken.chen@simagix.com", false)
	if isEmailAddress(email) == false {
		t.Fatal(email)
	}
}

func TestIsEmailAddress(t *testing.T) {
	email := "ken.chen@simagix.com"
	if isEmailAddress(email) == false {
		t.Fatal(email)
	}
	email = "ken.chen/simagix.com"
	if isEmailAddress(email) == true {
		t.Fatal(email)
	}

}

func TestGetEmailAddress(t *testing.T) {
	email := GetEmailAddress()
	if isEmailAddress(email) == false {
		t.Fatal(email)
	}
}

func TestIsIP(t *testing.T) {
	ip := "192.168.1.1"
	if isIP(ip) == false {
		t.Fatal(ip)
	}
	ip = "182.168.315.1"
	if isIP(ip) == true {
		t.Fatal(ip)
	}
}

func TestGetIP(t *testing.T) {
	ip := getIP()
	if isIP(ip) == false {
		t.Fatal(ip)
	}
}

func TestIsSSN(t *testing.T) {
	ssn := "555-12-3456"
	if isSSN(ssn) == false {
		t.Fatal(ssn)
	}
	ssn = "55-2343-0909"
	if isSSN(ssn) == true {
		t.Fatal(ssn)
	}
}

func TestGetSSN(t *testing.T) {
	ssn := getSSN()
	if isSSN(ssn) == false {
		t.Fatal(ssn)
	}
}

func TestIsPhoneNumber(t *testing.T) {
	phone := "(555) 123-4567"
	if isPhoneNumber(phone) == false {
		t.Fatal(phone)
	}
	phone = "+1886 2 555 1234"
	if isPhoneNumber(phone) == true {
		t.Fatal(phone)
	}
}

func TestGetPhoneNumber(t *testing.T) {
	phone := getPhoneNumber()
	if isPhoneNumber(phone) == false {
		t.Fatal(phone)
	}
}

func TestIsHexString(t *testing.T) {
	hex := "ABCDEF0123456789"
	if isHexString(hex) == false {
		t.Fatal(hex)
	}
	hex = "XXYYZZ"
	if isHexString(hex) == true {
		t.Fatal(hex)
	}
}

func TestGetHexString(t *testing.T) {
	hex := getHexString(8)
	if isHexString(hex) == false {
		t.Fatal(hex)
	}
}

func TestIsDateString(t *testing.T) {
	var dstr = "2018-10-15T12:00:00Z"
	if isDateString(dstr) == false {
		t.Fatal(dstr)
	}
	dstr = "10/15/2018 12:00:00"
	if isDateString(dstr) == true {
		t.Fatal(dstr)
	}
}

func TestGetDate(t *testing.T) {
	utime := getDate()
	t.Log(utime)
}
