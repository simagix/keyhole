// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"time"

	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
)

// NewMongoClient new mongo client
func NewMongoClient(uri string, opts ...string) (*mongo.Client, error) {
	var err error
	var client *mongo.Client
	if len(opts) >= 2 && opts[0] != "" && opts[1] != "" {
		var opt *options.ClientOptions
		opt = options.Client()
		var sslOpt = options.SSLOpt{Enabled: true, CaFile: opts[0], ClientCertificateKeyFile: opts[1]}
		opt.SetSSL(&sslOpt)
		if client, err = mongo.NewClientWithOptions(uri, opt); err != nil {
			return client, err
		}
	} else {
		if client, err = mongo.NewClient(uri); err != nil {
			return client, err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err = client.Connect(ctx); err != nil {
		panic(err)
	}
	return client, err
}
