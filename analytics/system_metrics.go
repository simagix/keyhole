// Copyright 2019 Kuei-chun Chen. All rights reserved.

package analytics

import (
	"time"
)

// SystemMetricsDoc -
type SystemMetricsDoc struct {
	Start time.Time              `json:"start" bson:"start"`
	CPU   CPUMetrics             `json:"cpu" bson:"cpu"`
	Disks map[string]DiskMetrics `json:"disks" bson:"disks"`
}

// CPUMetrics -
type CPUMetrics struct {
	IdleMS    uint64 `json:"idle_ms" bson:"idle_ms"`
	IOWaitMS  uint64 `json:"iowait_ms" bson:"iowait_ms"`
	NiceMS    uint64 `json:"nice_ms" bson:"nice_ms"`
	SoftirqMS uint64 `json:"softirq_ms" bson:"softirq_ms"`
	StealMS   uint64 `json:"steal_ms" bson:"steal_ms"`
	SystemMS  uint64 `json:"system_ms" bson:"system_ms"`
	UserMS    uint64 `json:"user_ms" bson:"user_ms"`
	TotalMS   uint64
}

// DiskMetrics -
type DiskMetrics struct {
	ReadTimeMS   uint64 `json:"read_time_ms" bson:"read_time_ms"`
	WriteTimeMS  uint64 `json:"write_time_ms" bson:"write_time_ms"`
	IOInProgress uint64 `json:"io_in_progress" bson:"io_in_progress"`
	IOQueuedMS   uint64 `json:"io_queued_ms" bson:"io_queued_ms"`
	IOTimeMS     uint64 `json:"io_time_ms" bson:"io_time_ms"`
	Reads        uint64 `json:"reads" bson:"reads"`
	Writes       uint64 `json:"writes" bson:"writes"`
}
