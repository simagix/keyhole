// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"testing"
)

func TestAnalyzeServerStatus(t *testing.T) {
	var filename = "/tmp/keyhole_stats.2018-10-18T080737-standalone.gz"
	var err error
	var docs []ServerStatusDoc

	if _, docs, err = AnalyzeServerStatus(filename); err != nil {
		t.Fatal(err)
	}
	t.Log(len(docs))
}
