// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
	"github.com/mongodb/mongo-go-driver/x/network/connstring"
	"golang.org/x/crypto/ssh/terminal"
)

// KEYHOLEDB -
const KEYHOLEDB = "_KEYHOLE_"

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

// Parse checks if password is included
func Parse(uri string) (string, error) {
	var err error
	var connString connstring.ConnString
	if connString, err = connstring.Parse(uri); err != nil {
		return uri, err
	}
	if connString.Username != "" && connString.Password == "" {
		if connString.Password, err = ReadPasswordFromStdin(); err != nil {
			return uri, err
		}
		index := strings.Index(uri, "@")
		uri = (uri)[:index] + ":" + connString.Password + (uri)[index:]
	}

	if connString.Database == "" {
		connString.Database = KEYHOLEDB
		pos := strings.Index(uri, "?")
		if pos > 0 { // found ?query_string
			uri = (uri)[:pos] + connString.Database + (uri)[pos:]
		} else {
			length := len(uri)
			if (uri)[length-1] == '/' {
				uri += connString.Database
			} else {
				uri += "/" + connString.Database
			}
		}
	}
	return uri, err
}

// ReadPasswordFromStdin reads password from stdin
func ReadPasswordFromStdin() (string, error) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		return "", errors.New("Missing password")
	}
	var buffer []byte
	var err error
	fmt.Print("Enter Password: ")
	if buffer, err = terminal.ReadPassword(int(syscall.Stdin)); err != nil {
		return "", err
	}
	return string(buffer), err
}
