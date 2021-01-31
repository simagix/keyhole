// Copyright 2021 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"errors"

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
func CollectCluserDetails(stats *mdb.ClusterStats, client *mongo.Client, connString connstring.ConnString) error {
	var err error
	if err = stats.GetClusterStats(client, connString); err != nil {
		return errors.New(`Roles 'clusterMonitor' and 'readAnyDatabase' are required`)
	}
	if err = stats.OutputBSON(); err != nil {
		return err
	}
	stats.Print()
	return err
}
