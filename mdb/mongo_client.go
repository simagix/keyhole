// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
)

// NewMongoClient new mongo client
func NewMongoClient(uri string, sslCAFile string, sslPEMKeyFile string) (*mongo.Client, error) {
	if sslCAFile == "" && sslPEMKeyFile == "" {
		return mongo.NewClient(uri)
	}
	var opt *options.ClientOptions
	opt = options.Client()
	var sslOpt = options.SSLOpt{Enabled: true, CaFile: sslCAFile, ClientCertificateKeyFile: sslPEMKeyFile}
	opt.SetSSL(&sslOpt)
	return mongo.NewClientWithOptions(uri, opt)
}
