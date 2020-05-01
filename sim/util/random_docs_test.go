// Copyright 2018 Kuei-chun Chen. All rights reserved.

package util

import (
	"reflect"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestGetRandomizedDoc(t *testing.T) {
	var err error
	var doc bson.M
	buf := []byte(`{"_id":"5d5eb4860560b066bd58127b","active":false,"array1":[6871,753,290],"array2":["leittl","ceut","glir"],"array3":[{"city1":"keo YwrN","city2":"altAatn","city3":"mMiia"},{"city1":"Cohacig","city2":"allaDs","city3":"uoHnots"}],"email":"Logan.R.Arthur@gmail.com","hex":"4132cbda","hostIP":"ATL","lastUpdated":"2019-10-16T21:04:30-04:00","longString":"etuingcvulYe. uoa  an rss rnh oials  aieTshmsnid enftyea ","number":619,"objectId":"5d5eb4860560b066bd58127a","phoneNumber":"(439) 555-8590","shortString":"atnaltA","ssn":"229-58-5408","subdocs":{"attribute1":{"email":"Ava.O.Smith@simagix.com"}}}`)

	if doc, err = GetRandomizedDoc(buf, false); err != nil {
		t.Fatal(err)
	}
	t.Log(doc)

	buf = []byte(`{"_id": Number(123), "xyz": NumberDecimal("456")}`)

	if doc, err = GetRandomizedDoc(buf, false); err != nil {
		t.Fatal(err)
	}
	t.Log(doc)
}

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

	_, ok := doc["numberLong"].(float64) // number data type
	if !ok {
		t.Fatal("expected int but got", reflect.TypeOf(doc["numberLong"]))
	}

	if doc, err = GetDocByTemplate("testdata/template.json", false); err != nil {
		t.Fatal(err)
	}

	_, ok = doc["_id"].(primitive.ObjectID)
	if !ok {
		t.Fatal("expected ObjectID but got", reflect.TypeOf(doc["_id"]))
	}
}

func TestGetMagicString(t *testing.T) {
	email := getMagicString("ken.chen@simagix.com", false)
	if isEmailAddress(email) == false {
		t.Fatal(email)
	}

	sn := getMagicString("Az0-1234-5678-9", true)
	if sn[0] > 90 || sn[1] > 122 || sn[2] > 57 || sn[3:4] != "-" {
		t.Fatal(sn)
	}

	uri := getMagicString("https://localhost:8080", true)
	if strings.HasPrefix(uri, "https://") == false {
		t.Fatal(uri)
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

func TestGetNumber(t *testing.T) {
	x := getNumber(int(123))
	if x.(int) < 100 || x.(int) >= 1000 {
		t.Fatal("expected between 100 and 1000, but got", x)
	}
	x = getNumber(int8(23))
	if x.(int8) < 10 || x.(int8) >= 100 {
		t.Fatal("expected between 100 and 1000, but got", x)
	}
	x = getNumber(int32(123))
	if x.(int32) < 100 || x.(int32) >= 1000 {
		t.Fatal("expected between 100 and 1000, but got", x)
	}
	x = getNumber(int64(123))
	if x.(int64) < 100 || x.(int64) >= 1000 {
		t.Fatal("expected between 100 and 1000, but got", x)
	}
	x = getNumber(float32(123))
	if x.(float32) < 100 || x.(float32) >= 1000 {
		t.Fatal("expected between 100 and 1000, but got", x)
	}
	x = getNumber(float64(123))
	if x.(float64) < 100 || x.(float64) >= 1000 {
		t.Fatal("expected between 100 and 1000, but got", x)
	}
}

func TestGetRandomNumber(t *testing.T) {
	x := getRandomNumber(float64(123))
	if x < 100 || x >= 1000 {
		t.Fatal("expected between 100 and 1000, but got", x)
	}
	x = getRandomNumber(float64(100))
	if x < 100 || x >= 1000 {
		t.Fatal("expected between 100 and 1000, but got", x)
	}
	x = getRandomNumber(float64(999))
	if x < 100 || x >= 1000 {
		t.Fatal("expected between 100 and 1000, but got", x)
	}
}
