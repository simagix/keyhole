// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
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

// GetReplLagsTSV -
func GetReplLagsTSV() []string {
	var docs []string
	var ts int

	str := "date"
	stat := ReplSetStatusDoc{}
	for i, doc := range ChartsDocs["replset"] {
		buf, _ := json.Marshal(doc)
		json.Unmarshal(buf, &stat)
		ts = 0
		sort.Slice(stat.Members, func(i, j int) bool { return stat.Members[i].Name < stat.Members[j].Name })
		if i == 0 {
			for _, mb := range stat.Members {
				a := strings.Index(mb.Name, ".")
				b := strings.LastIndex(mb.Name, ":")
				if a < 0 || b < 0 {
					str += "\t" + mb.Name
				} else {
					str += "\t" + mb.Name[0:a] + mb.Name[b:]
				}
			}
			docs = append(docs, str)
		}
		for _, mb := range stat.Members {
			if mb.StateStr == PRIMARY {
				ts = mb.Optime.TS
				break
			}
		}

		str = stat.Date.Format("2006-01-02T15:04:05Z")
		if ts == 0 {
			continue
		} else {
			for _, mb := range stat.Members {
				if mb.StateStr == SECONDARY {
					str += "\t" + strconv.Itoa((ts-mb.Optime.TS)/1000/1000/1000)
				} else if mb.StateStr == PRIMARY {
					str += "\t0"
				}
			}
			docs = append(docs, str)
		}
	}
	return docs
}
