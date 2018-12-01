// Copyright 2018 Kuei-chun Chen. All rights reserved.

package ftdc

// FTDC -
type FTDC struct {
	NumAttribs    uint32
	NumDeltas     uint32
	DataPointsMap map[string][]int64
}

// NewFTDC -
func NewFTDC() *FTDC {
	return &FTDC{DataPointsMap: map[string][]int64{}}
}
