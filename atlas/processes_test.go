// Copyright 2019 Kuei-chun Chen. All rights reserved.

package atlas

import (
	"os"
	"testing"

	"github.com/simagix/gox"
)

func TestGetProcesses(t *testing.T) {
	publicKey := os.Getenv("ATLAS_PUB")
	privateKey := os.Getenv("ATLAS_PRI")
	groupID := os.Getenv("ATLAS_GROUP")
	api := NewKey(publicKey, privateKey)
	api.SetVerbose(testing.Verbose())
	if doc, err := api.GetProcesses(groupID); err != nil {
		t.Fatal(err)
	} else {
		t.Log(gox.Stringify(doc, "", "  "))
	}
}
