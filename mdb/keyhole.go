// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

// Keyhole holds keyhole object info
type Keyhole struct {
	verbose bool
	version string
}

// NewKeyhole returns Keyhole structure
func NewKeyhole(version string) *Keyhole {
	keyhole := Keyhole{version: version}
	return &keyhole
}

// SetVerbose sets verbose mode
func (k *Keyhole) SetVerbose(verbose bool) {
	k.verbose = verbose
}

// GetClusterSummary returns one line cluster summary
func (k *Keyhole) GetClusterSummary(client *mongo.Client) string {
	stats := NewStats(k.version)
	stats.SetVerbose(k.verbose)
	return stats.GetClusterShortSummary(client)
}

// CollectClusterStats collects cluster stats
func (k *Keyhole) CollectClusterStats(client *mongo.Client, connString connstring.ConnString) (string, error) {
	var err error
	var cluster ClusterDetails
	stats := NewStats(k.version)
	stats.SetVerbose(k.verbose)
	if cluster, err = stats.GetClusterStats(client, connString); err != nil {
		return "", err
	}
	return stats.Save(cluster)
}
