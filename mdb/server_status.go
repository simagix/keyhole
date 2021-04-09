// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// ServerStatus stores build information
type ServerStatus struct {
	Connections struct {
		Active       int `bson:"active"`
		Current      int `bson:"current"`
		Available    int `bson:"available"`
		TotalCreated int `bson:"totalCreated"`
	} `bson:"connections"`
	Host      string    `bson:"host"`
	LocalTime time.Time `bson:"localTime"`
	Mem       struct {
		Virtual   int  `bson:"virtual"`
		Supported bool `bson:"supported"`
		Bits      int  `bson:"bits"`
		Resident  int  `bson:"resident"`
	} `bson:"mem"`
	Metrics struct {
		Cursor struct {
			Open struct {
				Pinned int `bson:"pinned"`
				Total  int `bson:"total"`
			} `bson:"open"`
		} `bson:"cursor"`
	} `bson:"metrics"`
	OpCounters struct {
		Command int `bson:"command"`
		Insert  int `bson:"insert"`
		Query   int `bson:"query"`
		Update  int `bson:"update"`
		Delete  int `bson:"delete"`
		GetMore int `bson:"getmore"`
	} `bson:"opcounters"`
	Process string `bson:"process"`
	Repl    struct {
		Hosts     []string `bson:"hosts"`
		IsMaster  bool     `bson:"isMaster"`
		Me        string   `bson:"me"`
		Primary   string   `bson:"primary"`
		SetName   string   `bson:"setName"`
		Secondary bool     `bson:"secondary"`
	} `bson:"repl"`
	Sharding struct {
		ConfigsvrConnectionString string `bson:"configsvrConnectionString"`
		MaxChunkSizeInBytes       int    `bson:"maxChunkSizeInBytes"`
	} `bson:"sharding"`
	StorageEngine struct {
		Name string `bson:"name"`
	} `bson:"storageEngine"`
	Version    string `bson:"version"`
	WiredTiger struct {
		Cache bson.M `bson:"cache"`
	} `bson:"wiredTiger"`
}

// GetServerStatus returns MongoDB build information
func GetServerStatus(client *mongo.Client) (ServerStatus, error) {
	ctx := context.Background()
	var serverStatus ServerStatus
	err := client.Database("admin").RunCommand(ctx, bson.D{{Key: "serverStatus", Value: 1}}).Decode(&serverStatus)
	return serverStatus, err
}

// GetClusterType returns cluster type
func GetClusterType(serverStatus ServerStatus) string {
	if serverStatus.Repl.SetName != "" {
		return Replica
	} else if serverStatus.Sharding.ConfigsvrConnectionString != "" {
		return Sharded
	}
	return Standalone
}
