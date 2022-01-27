// Copyright 2019 Kuei-chun Chen. All rights reserved.

package atlas

import (
	"os"
	"testing"
)

func TestGetClustersSummary(t *testing.T) {
	var err error
	var str string

	publicKey := os.Getenv("ATLAS_USER_PS")
	privateKey := os.Getenv("ATLAS_KEY_PS")
	api := NewKey(publicKey, privateKey)
	// api.SetVerbose(testing.Verbose())
	if str, err = api.GetClustersSummary(); err != nil {
		t.Fatal(err)
	}
	t.Log(str)
}
