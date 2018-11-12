// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"time"
)

// PRIMARY - primary node
const PRIMARY = "PRIMARY"

// SECONDARY - secondary node
const SECONDARY = "SECONDARY"

// OptimeDoc -
type OptimeDoc struct {
	T  int `json:"t" bson:"t"`
	TS int `json:"ts" bson:"ts"`
}

// MemberDoc stores replset status
type MemberDoc struct {
	Name     string    `json:"name" bson:"name"`
	Optime   OptimeDoc `json:"optime" bson:"optime"`
	StateStr string    `json:"stateStr" bson:"stateStr"`
}

// ReplSetStatusDoc stores replset status
type ReplSetStatusDoc struct {
	Date    time.Time   `json:"date" bson:"date"`
	Members []MemberDoc `json:"members" bson:"members"`
}
