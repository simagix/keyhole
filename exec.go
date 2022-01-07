// Copyright 2020-present Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"fmt"
	"io/ioutil"

	"go.mongodb.org/mongo-driver/bson"
)

const (
	printConnections = "print_connections"
)

// Config stores keyhole configuration
type Config struct {
	Action string `bson:"action,omitempty"`
	URI    string `bson:"uri,omitempty"`
}

// Exec executes a plan based on a configuration file
func Exec(filename string) error {
	var err error
	var cfg *Config
	var data []byte
	if data, err = ioutil.ReadFile(filename); err != nil {
		return err
	} else if err = bson.UnmarshalExtJSON(data, false, &cfg); err != nil {
		return err
	} else if cfg.Action == "" {
		return fmt.Errorf(`action is required`)
	}

	if cfg.Action == printConnections {
		return PrintConnectionsFromURI(cfg.URI)
	}
	return err
}
