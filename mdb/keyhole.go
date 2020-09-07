// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

// Keyhole holds keyhole object info
type Keyhole struct {
	redaction bool
	verbose   bool
	version   string
}

// NewKeyhole returns Keyhole structure
func NewKeyhole(version string) *Keyhole {
	keyhole := Keyhole{version: version}
	return &keyhole
}

// SetRedaction sets redaction
func (k *Keyhole) SetRedaction(redaction bool) {
	k.redaction = redaction
}

// SetVerbose sets verbose mode
func (k *Keyhole) SetVerbose(verbose bool) {
	k.verbose = verbose
}

// GetClusterSummary returns one line cluster summary
func (k *Keyhole) GetClusterSummary(client *mongo.Client) string {
	stats := NewStats(k.version)
	stats.SetRedaction(k.redaction)
	stats.SetVerbose(k.verbose)
	return stats.GetClusterShortSummary(client)
}

// CollectClusterStats collects cluster stats
func (k *Keyhole) CollectClusterStats(client *mongo.Client, connString connstring.ConnString) error {
	var err error
	stats := NewStats(k.version)
	stats.SetRedaction(k.redaction)
	stats.SetVerbose(true)
	if err = stats.GetClusterStats(client, connString); err != nil {
		result := `Roles 'clusterMonitor' and 'readAnyDatabase' are required`
		log.Println(result)
		return err
	}
	if err = stats.OutputBSON(); err != nil {
		return err
	}
	stats.Print()
	return err
}
