// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

type SystemMetricsDoc struct {
	Start time.Time  `json:"start" bson:"start"`
	CPU   CPUMetrics `json:"cpu" bson:"cpu"`
	Disks bson.M     `json:"disks" bson:"disks"`
}

type CPUMetrics struct {
	IdleMS    int `json:"idle_ms" bson:"idle_ms"`
	IOWaitMS  int `json:"iowait_ms" bson:"iowait_ms"`
	NiceMS    int `json:"nice_ms" bson:"nice_ms"`
	SoftirqMS int `json:"softirq_ms" bson:"softirq_ms"`
	StealMS   int `json:"steal_ms" bson:"steal_ms"`
	SystemMS  int `json:"system_ms" bson:"system_ms"`
	UserMS    int `json:"user_ms" bson:"user_ms"`
}

type DiskMetrics struct {
	ReadTimeMS  int `json:"read_time_ms" bson:"read_time_ms"`
	WriteTimeMS int `json:"write_time_ms" bson:"write_time_ms"`
	IOTimeMS    int `json:"io_time_ms" bson:"io_time_ms"`
}
