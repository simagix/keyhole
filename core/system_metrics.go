// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import "time"

type SystemMetricsDoc struct {
	Start time.Time  `json:"start" bson:"start"`
	CPU   CPUMetrics `json:"cpu" bson:"cpu"`
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
