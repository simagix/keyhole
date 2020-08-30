// Copyright 2018 Kuei-chun Chen. All rights reserved.

package ftdc

// MetricsData -
type MetricsData struct {
	Block         []byte
	DataPointsMap map[string][]uint64
	NumDeltas     uint32
}

// Metrics -
type Metrics struct {
	Doc  interface{}   // type 0
	Data []MetricsData // type 1
}

// NewMetrics -
func NewMetrics() *Metrics {
	return &Metrics{}
}
