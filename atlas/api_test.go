// Copyright 2019 Kuei-chun Chen. All rights reserved.

package atlas

import (
	"os"
	"testing"
)

func TestNewKey(t *testing.T) {
	publicKey := os.Getenv("ATLAS_PUB")
	privateKey := os.Getenv("ATLAS_PRI")
	api := NewKey(publicKey, privateKey)
	if api.publicKey != publicKey || api.privateKey != privateKey {
		t.Fatal("parsing error")
	}
}

func TestParseURI(t *testing.T) {
	var err error
	var api *API
	publicKey := os.Getenv("ATLAS_PUB")
	privateKey := os.Getenv("ATLAS_PRI")
	if api, err = ParseURI(os.Getenv("ATLAS_AUTH")); err != nil {
		t.Fatal(err)
	}
	if api.publicKey != publicKey || api.privateKey != privateKey {
		t.Fatal(publicKey, privateKey, os.Getenv("ATLAS_AUTH"))
	}
}

func TestSetVerbose(t *testing.T) {
	publicKey := os.Getenv("ATLAS_PUB")
	privateKey := os.Getenv("ATLAS_PRI")
	api := NewKey(publicKey, privateKey)
	api.SetVerbose(testing.Verbose())
	if api.verbose != testing.Verbose() {
		t.Fatal("SetVerbose failed")
	}
}

func TestGET(t *testing.T) {
	publicKey := os.Getenv("ATLAS_PUB")
	privateKey := os.Getenv("ATLAS_PRI")
	groupID := os.Getenv("ATLAS_GROUP")
	api := NewKey(publicKey, privateKey)
	uri := BaseURL + "/groups/" + groupID + "/clusters"
	if _, err := api.Get(uri); err != nil {
		t.Fatal(err)
	}
}
