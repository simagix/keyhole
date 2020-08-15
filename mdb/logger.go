// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"fmt"
	"log"
	"strings"
	"time"
)

// Logger stores logger info
type Logger struct {
	Collected time.Time
	Logs      []string
	Params    string
	Version   string
}

// NewLogger returns Logger
func NewLogger(version string, params string) *Logger {
	k := Logger{Version: version, Params: params}
	now := time.Now()
	k.Collected = now
	k.Logs = []string{fmt.Sprintf(`%v keyhole begins`, now.Format(time.RFC3339))}
	return &k
}

// Add adds a message
func (p *Logger) Add(message string) {
	str := fmt.Sprintf(`%v %v`, time.Now().Format(time.RFC3339), message)
	p.Logs = append(p.Logs, str)
}

// Log adds and prints a message
func (p *Logger) Log(message string) {
	str := fmt.Sprintf(`%v %v`, time.Now().Format(time.RFC3339), message)
	p.Logs = append(p.Logs, str)
	log.Println(message)
}

// Print prints keyhole info
func (p *Logger) Print() string {
	if p == nil {
		return ""
	}
	strs := []string{fmt.Sprintf(`{ keyhole: { version: "%v", args: "%v" } }`, p.Version, p.Params)}
	strs = append(strs, p.Logs...)
	return strings.Join(strs, "\n")
}
