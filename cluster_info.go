// Copyright 2021 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"github.com/simagix/keyhole/mdb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

// GetClusterSummary returns MongoDB cluster summary
func GetClusterSummary(version string, client *mongo.Client) string {
	stats := mdb.NewClusterStats(version)
	return stats.GetClusterShortSummary(client)
}

// CollectCluserDetails collects cluster details
func CollectCluserDetails(stats *mdb.ClusterStats, client *mongo.Client, connString connstring.ConnString, maobiURL string) error {
	var err error
	var data []byte
	var ofile string
	if err = stats.GetClusterStats(client, connString); err != nil {
		return err
	}
	stats.Print()
	if ofile, data, err = stats.OutputBSON(); err != nil {
		return err
	}
	return GenerateMaobiReport(maobiURL, data, ofile)
}
