// Copyright 2019 Kuei-chun Chen. All rights reserved.

package analytics

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

const DiagnosticDataDirectory = "../diagnostic.data"
const DiagnosticDataFilename = DiagnosticDataDirectory + "/metrics.2017-10-12T20-08-53Z-00000"

func TestReadDiagnosticFiles(t *testing.T) {
	var err error
	var files []os.FileInfo
	var filenames []string

	if files, err = ioutil.ReadDir(DiagnosticDataDirectory); err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		if strings.Index(f.Name(), "metrics.") != 0 && strings.Index(f.Name(), "keyhole_stats.") != 0 {
			continue
		}
		filename := DiagnosticDataDirectory + "/" + f.Name()
		filenames = append(filenames, filename)
	}
	d := NewDiagnosticData()
	if err = d.readDiagnosticFiles(filenames); err != nil {
		t.Fatal(err)
	}
}
