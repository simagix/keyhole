// Copyright 2020-present Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"fmt"
	"os"

	"github.com/simagix/gox"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	compareClusters  = "compare_clusters"
	printConnections = "print_connections"
)

// Config stores keyhole configuration
type Config struct {
	Action    string `bson:"action,omitempty"`
	Filename  string `bson:"filename,omitempty"`
	NoColor   bool   `bson:"no_color,omitempty"`
	Signature string `bson:"signature,omitempty"`
	URI       string `bson:"uri,omitempty"`
	Verbose   bool   `bson:"verbose,omitempty"`

	IsDeepCompare bool     `bson:"deep_compare,omitempty"`
	Filters       []Filter `bson:"filters,omitempty"`
	SampleSize    int      `bson:"sample_size,omitempty"`
	SourceURI     string   `bson:"source_uri,omitempty"`
	TargetURI     string   `bson:"target_uri,omitempty"`
}

// Filter holds query filter of a namespace
type Filter struct {
	NS       string `bson:"ns"`
	Query    bson.D `bson:"query"`
	TargetNS string `bson:"target_ns,omitempty"`
}

// Exec executes a plan based on a configuration file
func Exec(filename string, signature string) error {
	var err error
	var cfg *Config
	var data []byte
	if data, err = os.ReadFile(filename); err != nil {
		return err
	} else if err = bson.UnmarshalExtJSON(data, false, &cfg); err != nil {
		return err
	} else if cfg.Action == "" {
		return fmt.Errorf(`action is required`)
	}
	if cfg.Signature == "" {
		cfg.Signature = signature
	}
	gox.GetLogger(cfg.Signature).Infof("executing %v", cfg.Action)
	if cfg.Action == compareClusters {
		return CompareClusters(cfg)
	} else if cfg.Action == printConnections {
		return PrintConnections(cfg)
	}
	return err
}
