// Copyright 2019 Kuei-chun Chen. All rights reserved.

package atlas

import (
	"os"
	"testing"
	"time"
)

func TestDownloadLogs(t *testing.T) {
	var err error
	atlas := ParseAtlasURI("atlas://" + os.Getenv("ATLAS_AUTH") + "@" + os.Getenv("ATLAS_GROUP") + "/keyhole")
	if _, err = atlas.DownloadLogs(os.TempDir()); err != nil {
		t.Fatal(err)
	}
}

func TestDownloadLog(t *testing.T) {
	var err error
	atlas := ParseAtlasURI("atlas://" + os.Getenv("ATLAS_AUTH") + "@" + os.Getenv("ATLAS_GROUP") + "/keyhole")
	startUnix := time.Now().Add(time.Hour * -24 * 3).Unix()
	endUnix := time.Now().Unix()
	if _, err = atlas.downloadLog(os.TempDir(), "keyhole-shard-00-00-jgtm2.mongodb.net", startUnix, endUnix); err != nil {
		t.Fatal(err)
	}
}
