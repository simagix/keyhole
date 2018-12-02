// Copyright 2018 Kuei-chun Chen. All rights reserved.

package ftdc

import "bytes"

// Metrics -
type Metrics struct {
	Blocks []bytes.Buffer
	Doc    interface{}   // type 0
	Data   []MetricsData // type 1
}

// MetricsData -
type MetricsData struct {
	DataSize      uint32
	NumAttribs    uint32
	NumDeltas     uint32
	DataPointsMap map[string][]int64
}

// NewMetrics -
func NewMetrics() *Metrics {
	return &Metrics{}
}
