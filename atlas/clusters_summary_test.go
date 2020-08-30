// Copyright 2019 Kuei-chun Chen. All rights reserved.

package atlas

import (
	"os"
	"testing"
)

func TestGetClustersSummary(t *testing.T) {
	publicKey := os.Getenv("ATLAS_PUB")
	privateKey := os.Getenv("ATLAS_PRI")
	api := NewKey(publicKey, privateKey)
	api.SetVerbose(testing.Verbose())
	if str, err := api.GetClustersSummary(); err != nil {
		t.Fatal(err)
	} else {
		t.Log(str)
	}
}
