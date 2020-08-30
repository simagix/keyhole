// Copyright 2019 Kuei-chun Chen. All rights reserved.

package analytics

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"time"
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
				if file.IsDir() == false &&
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

func parseTime(filename string) (time.Time, error) {
	layout := "2006-01-02T15-04-05Z"
	x := strings.Index(filename, "metrics.")
	y := strings.LastIndex(filename, "-")
	if x < 0 || y < 0 || y < x {
		return time.Now(), errors.New("not valid")
	}
	t, err := time.Parse(layout, filename[x+8:y])
	if err != nil {
		return time.Now(), err
	}
	return t, nil
}
