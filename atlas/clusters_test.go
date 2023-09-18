// Copyright 2019 Kuei-chun Chen. All rights reserved.

package atlas

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestGetClusters(t *testing.T) {
	var err error
	var data []byte
	var doc map[string]interface{}

	publicKey := os.Getenv("ATLAS_USER_PS")
	privateKey := os.Getenv("ATLAS_KEY_PS")
	groupID := os.Getenv("ATLAS_GROUP")
	api := NewKey(publicKey, privateKey)
	api.SetVerbose(testing.Verbose())
	if doc, err = api.GetClusters(groupID); err != nil {
		t.Fatal(err)
	}
	if data, err = json.Marshal(doc); err != nil {
		t.Fatal(err)
	}
	os.Mkdir("./out", 0755)
	ofile := `./out/clusters.json`
	if err = os.WriteFile(ofile, data, 0644); err != nil {
		t.Fatal(err)
	}
}

func TestGetCluster(t *testing.T) {
	var err error
	var data []byte
	var doc map[string]interface{}
	clusterName := "demo"

	publicKey := os.Getenv("ATLAS_USER_PS")
	privateKey := os.Getenv("ATLAS_KEY_PS")
	groupID := os.Getenv("ATLAS_GROUP")
	api := NewKey(publicKey, privateKey)
	// api.SetVerbose(testing.Verbose())
	if doc, err = api.GetCluster(groupID, clusterName); err != nil {
		t.Fatal(err)
	}
	if data, err = json.Marshal(doc); err != nil {
		t.Fatal(err)
	}
	os.Mkdir("./out", 0755)
	ofile := fmt.Sprintf(`./out/cluster-%v.json`, clusterName)
	if err = os.WriteFile(ofile, data, 0644); err != nil {
		t.Fatal(err)
	}
}
