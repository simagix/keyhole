// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"testing"
)

func TestParseDialInfo(t *testing.T) {
	uri := "mongodb://root:secret@keyhole-00-jgtm2.mongodb.net,keyhole-01-jgtm2.mongodb.net/test"
	dialInfo, err := ParseDialInfo(uri)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(dialInfo.Addrs)
}

func TestParseDialInfoSrv(t *testing.T) {
	uri := "mongodb+srv://root:secret@keyhole-jgtm2.mongodb.net/test"
	dialInfo, err := ParseDialInfo(uri)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(dialInfo.Addrs)
}
