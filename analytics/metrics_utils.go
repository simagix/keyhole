// Copyright 2019 Kuei-chun Chen. All rights reserved.

package analytics

import (
	"io/ioutil"
	"os"
	"strings"
)

func getFilenames(filenames []string) []string {
	var err error
	var fi os.FileInfo
	fnames := []string{}
	for _, filename := range filenames {
		if fi, err = os.Stat(filename); err != nil {
			continue
		}
		switch mode := fi.Mode(); {
		case mode.IsDir():
			files, _ := ioutil.ReadDir(filename)
			for _, file := range files {
				if file.IsDir() &&
					(strings.HasPrefix(file.Name(), "metrics.") || strings.HasPrefix(file.Name(), "keyhole_stats.")) {
					fnames = append(fnames, filename+"/"+file.Name())
				}
			}
		case mode.IsRegular():
			fnames = append(fnames, filename)
		}
	}
	return fnames
}
