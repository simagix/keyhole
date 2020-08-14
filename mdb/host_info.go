// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// HostInfo stores build information
type HostInfo struct {
	Extra struct {
		KernelVersion string
	}
	OS struct {
		Name    string
		Type    string
		Version string
	}
	System struct {
		CPUAddrSize int    `bson:"cpuAddrSize,truncate"`
		CPUArch     string `bson:"cpuArch"`
		Hostname    string `bson:"hostname"`
		MemLimitMB  int    `bson:"memLimitMB,truncate"`
		MemSizeMB   int    `bson:"memSizeMB,truncate"`
		NumaEnabled bool   `bson:"numaEnabled"`
		NumCores    int    `bson:"numCores,truncate"`
	}
}

// GetHostInfo returns MongoDB build information
func GetHostInfo(client *mongo.Client) (HostInfo, error) {
	ctx := context.Background()
	var hostInfo HostInfo
	err := client.Database("admin").RunCommand(ctx, bson.D{{Key: "hostInfo", Value: 1}}).Decode(&hostInfo)
	return hostInfo, err
}
