// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"fmt"
	"log"
	"strings"
	"time"
)

// Audit stores keyhole info
type Audit struct {
	Collected time.Time
	Logs      []string
	Params    string
	Version   string
}

// NewAudit returns Audit
func NewAudit(version string, params string) *Audit {
	audit := Audit{Version: version, Params: params}
	now := time.Now()
	audit.Collected = now
	audit.Logs = []string{fmt.Sprintf(`%v keyhole begins`, now.Format(time.RFC3339))}
	return &audit
}

// Add adds a message
func (p *Audit) Add(message string) {
	str := fmt.Sprintf(`%v %v`, time.Now().Format(time.RFC3339), message)
	p.Logs = append(p.Logs, str)
}

// Log adds and prints a message
func (p *Audit) Log(message string) {
	str := fmt.Sprintf(`%v %v`, time.Now().Format(time.RFC3339), message)
	p.Logs = append(p.Logs, str)
	log.Println(message)
}

// Print prints keyhole info
func (p *Audit) Print() string {
	if p == nil {
		return ""
	}
	strs := []string{fmt.Sprintf(`{ keyhole: { version: "%v", args: "%v" } }`, p.Version, p.Params)}
	strs = append(strs, p.Logs...)
	return strings.Join(strs, "\n")
}
