// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

// SystemMetricsDoc -
type SystemMetricsDoc struct {
	Start time.Time  `json:"start" bson:"start"`
	CPU   CPUMetrics `json:"cpu" bson:"cpu"`
	Disks bson.M     `json:"disks" bson:"disks"`
}

// CPUMetrics -
type CPUMetrics struct {
	IdleMS    int64 `json:"idle_ms" bson:"idle_ms"`
	IOWaitMS  int64 `json:"iowait_ms" bson:"iowait_ms"`
	NiceMS    int64 `json:"nice_ms" bson:"nice_ms"`
	SoftirqMS int64 `json:"softirq_ms" bson:"softirq_ms"`
	StealMS   int64 `json:"steal_ms" bson:"steal_ms"`
	SystemMS  int64 `json:"system_ms" bson:"system_ms"`
	UserMS    int64 `json:"user_ms" bson:"user_ms"`
	TotalMS   int64
}

// DiskMetrics -
type DiskMetrics struct {
	ReadTimeMS  int64 `json:"read_time_ms" bson:"read_time_ms"`
	WriteTimeMS int64 `json:"write_time_ms" bson:"write_time_ms"`
	IO          int64
	IOQueuedMS  int64 `json:"io_queued_ms" bson:"io_queued_ms"`
	IOTimeMS    int64 `json:"io_time_ms" bson:"io_time_ms"`
	Reads       int64 `json:"reads" bson:"reads"`
	Writes      int64 `json:"writes" bson:"writes"`
}
