// Copyright 2019 Kuei-chun Chen. All rights reserved.

package atlas

import (
	"os"
	"testing"
)

func TestDownload(t *testing.T) {
	var err error
	var api *API
	var filenames []string
	publicKey := os.Getenv("ATLAS_PUB")
	privateKey := os.Getenv("ATLAS_PRI")
	groupID := os.Getenv("ATLAS_GROUP")
	uri := "atlas://" + publicKey + ":" + privateKey + "@" + groupID + "/keyhole"
	if api, err = ParseURI(uri); err != nil {
		t.Fatal(err)
	}
	api.SetVerbose(testing.Verbose())
	if filenames, err = api.DownloadLogs(); err != nil {
		t.Fatal(err)
	}
	for _, filename := range filenames {
		os.Remove(filename)
	}
}
