// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"strings"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
	"golang.org/x/crypto/ssh/terminal"
)

// KEYHOLEDB -
const KEYHOLEDB = "_KEYHOLE_"

// NewMongoClient new mongo client
func NewMongoClient(uri string, files ...string) (*mongo.Client, error) {
	var err error
	var client *mongo.Client
	var connString connstring.ConnString

	if uri, err = parse(uri); err != nil {
		return client, err
	}
	if connString, err = connstring.Parse(uri); err != nil {
		return client, err
	}
	opts := options.Client().ApplyURI(uri)
	if connString.ReplicaSet == "" {
		opts.SetDirect(true)
	}
	if connString.Username == "" {
		opts.Auth = nil
	}
	if len(files) > 0 && files[0] != "" {
		connString.SSL = true
		roots := x509.NewCertPool()
		var caBytes []byte
		if caBytes, err = ioutil.ReadFile(files[0]); err != nil {
			return nil, err
		}
		if ok := roots.AppendCertsFromPEM(caBytes); !ok {
			return client, errors.New("failed to parse root certificate")
		}
		var certs tls.Certificate
		if len(files) >= 2 && files[1] != "" {
			var clientBytes []byte
			if clientBytes, err = ioutil.ReadFile(files[1]); err != nil {
				return nil, err
			}
			if certs, err = tls.X509KeyPair(clientBytes, clientBytes); err != nil {
				return nil, err
			}
		}
		opts.SetTLSConfig(&tls.Config{RootCAs: roots, Certificates: []tls.Certificate{certs}})
	}
	if client, err = mongo.NewClient(opts); err != nil {
		return client, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err = client.Connect(ctx); err != nil {
		panic(err)
	}
	err = client.Ping(ctx, nil)
	return client, err
}

// parse checks if password is included
func parse(uri string) (string, error) {
	var err error
	var connString connstring.ConnString
	if connString, err = connstring.Parse(uri); err != nil {
		return uri, err
	}
	if connString.Username != "" && connString.Password == "" {
		if connString.Password, err = readPasswordFromStdin(); err != nil {
			return uri, err
		}
		index := strings.LastIndex(uri, "@")
		uri = (uri)[:index] + ":" + template.URLQueryEscaper(connString.Password) + (uri)[index:]
	}

	if connString.Database == "" {
		connString.Database = KEYHOLEDB
		pos := strings.LastIndex(uri, "?")
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

// readPasswordFromStdin reads password from stdin
func readPasswordFromStdin() (string, error) {
	var buffer []byte
	var err error
	fmt.Print("Enter Password: ")
	if buffer, err = terminal.ReadPassword(int(syscall.Stdin)); err != nil {
		return "", err
	}
	return string(buffer), err
}
