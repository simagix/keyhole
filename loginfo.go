// Copyright 2021 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"fmt"

	"github.com/simagix/keyhole/mdb"
)

// AnalyzeMongoLogs a helper function to analyze logs
func AnalyzeMongoLogs(loginfo *mdb.LogInfo, filenames []string, maobiURL string) error {
	var err error
	var data []byte
	var ofile string
	for _, filename := range filenames {
		if err = loginfo.AnalyzeFile(filename); err != nil {
			fmt.Println(err)
			continue
		}
		loginfo.Print()
		if ofile, data, err = loginfo.OutputBSON(); err != nil {
			fmt.Println(err)
			continue
		}
		if err = GenerateMaobiReport(maobiURL, data, ofile); err != nil {
			fmt.Println(err)
		}
	}
	return err
}
