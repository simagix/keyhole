// Copyright 2018 Kuei-chun Chen. All rights reserved.

package ftdc

// Block - buffer holder
type Block []byte

// MetricsData -
type MetricsData struct {
	Buffer        Block
	DocSize       uint32
	NumAttribs    uint32
	NumDeltas     uint32
	DataPointsMap map[string][]int64
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
