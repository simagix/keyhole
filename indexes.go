// Copyright 2021 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"github.com/simagix/keyhole/mdb"
	"go.mongodb.org/mongo-driver/mongo"
)

// DuplicateIndexesFromFile duplicate indexes from definitions defined in a file
func DuplicateIndexesFromFile(index *mdb.IndexStats, client *mongo.Client, filename string, isDrop bool) error {
	var err error
	var databases []mdb.Database
	if err = index.SetClusterDetailsFromFile(filename); err != nil {
		return err
	}
	if err = index.CopyIndexes(client, isDrop); err != nil {
		return err
	}
	if databases, err = index.GetIndexes(client); err != nil {
		return err
	}
	index.PrintIndexesOf(databases)
	return err
}

// CollectIndexStats collects all indexes stats
func CollectIndexStats(index *mdb.IndexStats, client *mongo.Client, maobiURL string) error {
	var err error
	var data []byte
	var ofile string
	var databases []mdb.Database
	if databases, err = index.GetIndexes(client); err != nil {
		return err
	}
	index.PrintIndexesOf(databases)
	if ofile, data, err = index.OutputBSON(); err != nil {
		return err
	}
	return GenerateMaobiReport(maobiURL, data, ofile)
}
