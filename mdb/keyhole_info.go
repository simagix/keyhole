// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

// KeyholeInfo stores keyhole info
type KeyholeInfo struct {
	Collected time.Time
	Logs      []string
	Params    string
	Version   string
}

// NewKeyholeInfo returns KeyholeInfo
func NewKeyholeInfo(version string, params string) *KeyholeInfo {
	keyholeInfo := KeyholeInfo{Version: version, Params: params}
	now := time.Now()
	keyholeInfo.Collected = now
	keyholeInfo.Logs = []string{fmt.Sprintf(`%v keyhole begins`, now.Format(time.RFC3339))}
	return &keyholeInfo
}

// Log adds a message
func (ptr *KeyholeInfo) Log(s string) {
	ptr.Logs = append(ptr.Logs, fmt.Sprintf(`%v %v`, time.Now().Format(time.RFC3339), s))
}

// Print prints keyhole info
func (ptr *KeyholeInfo) Print() string {
	if ptr == nil {
		return ""
	}
	strs := []string{fmt.Sprintf(`{ keyhole: { version: "%v", args: "%v" } }`, ptr.Version, ptr.Params)}
	strs = append(strs, ptr.Logs...)
	return strings.Join(strs, "\n")
}

// GetClusterShortSummary is a wrapper of NewStats.GetClusterShortSummary
func (ptr *KeyholeInfo) GetClusterShortSummary(client *mongo.Client) string {
	var err error
	var cluster ClusterDetails
	stats := NewStats(ptr.Version)
	if cluster, err = stats.GetClusterStatsSummary(client); err != nil {
		return err.Error()
	}
	return stats.GetClusterShortSummary(cluster)
}
