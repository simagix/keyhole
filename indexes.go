// Copyright 2021 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"github.com/simagix/keyhole/mdb"
	"go.mongodb.org/mongo-driver/mongo"
)

// DuplicateIndexesFromFile duplicate indexes from definitions defined in a file
func DuplicateIndexesFromFile(index *mdb.IndexStats, client *mongo.Client, filename string) error {
	var err error
	var databases []mdb.Database
	if err = index.SetClusterDetailsFromFile(filename); err != nil {
		return err
	}
	if err = index.CreateIndexes(client); err != nil {
		return err
	}
	if databases, err = index.GetIndexes(client); err != nil {
		return err
	}
	index.PrintIndexesOf(databases)
	return err
}

// CollectIndexStats collects all indexes stats
func CollectIndexStats(index *mdb.IndexStats, client *mongo.Client) error {
	var err error
	var databases []mdb.Database
	if databases, err = index.GetIndexes(client); err != nil {
		return err
	}
	index.PrintIndexesOf(databases)
	if err = index.OutputBSON(); err != nil {
		return err
	}
	return err
}
