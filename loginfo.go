// Copyright 2021 Kuei-chun Chen. All rights reserved.

package keyhole

import "github.com/simagix/keyhole/mdb"

// AnalyzeMongoLogs a helper function to analyze logs
func AnalyzeMongoLogs(loginfo *mdb.LogInfo, filenames []string) error {
	var err error
	for _, filename := range filenames {
		if err = loginfo.AnalyzeFile(filename); err != nil {
			return err
		}
		loginfo.Print()
		loginfo.OutputBSON()
	}
	return err
}
