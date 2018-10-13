// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"encoding/json"
	"testing"
)

func TestParseDialInfo(t *testing.T) {
	uri := "mongodb://root:secret@keyhole-00-jgtm2.mongodb.net,keyhole-01-jgtm2.mongodb.net/test"
	dialInfo, err := ParseDialInfo(uri)
	if err != nil {
		t.Fatal(err)
	}
	bytes, _ := json.MarshalIndent(dialInfo.Addrs, "", "  ")
	t.Log(string(bytes))
}

func TestParseDialInfoSrv(t *testing.T) {
	uri := "mongodb+srv://root:secret@keyhole-jgtm2.mongodb.net/test"
	dialInfo, err := ParseDialInfo(uri)
	if err != nil {
		t.Fatal(err)
	}
	bytes, _ := json.MarshalIndent(dialInfo.Addrs, "", "  ")
	t.Log(string(bytes))
}

func TestNewReader(t *testing.T) {
}

func TestCountLines(t *testing.T) {
}
