// Copyright 2019 Kuei-chun Chen. All rights reserved.

package atlas

import (
	"os"
	"testing"
)

func TestPrintSummary(t *testing.T) {
	var err error
	var str string
	atlas := NewSummary(os.Getenv("ATLAS_AUTH"))

	if str, err = atlas.GetSummary(); err != nil {
		t.Fatal(err)
	}

	t.Log(str)
}

func TestGetRolesByName(t *testing.T) {
	atlas := NewSummary(os.Getenv("ATLAS_AUTH"))

	if _, err := atlas.getRolesByName(); err != nil {
		t.Fatal(err)
	}
}

func TestGetClusters(t *testing.T) {
	atlas := NewSummary(os.Getenv("ATLAS_AUTH"))

	if _, err := atlas.getClusters(os.Getenv("ATLAS_GROUP")); err != nil {
		t.Fatal(err)
	}
}

func TestGetProcesses(t *testing.T) {
	atlas := NewSummary(os.Getenv("ATLAS_AUTH"))

	if _, err := atlas.getProcesses(os.Getenv("ATLAS_GROUP")); err != nil {
		t.Fatal(err)
	}
}
