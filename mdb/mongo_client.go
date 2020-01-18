// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
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
func NewMongoClient(uri string, opts ...string) (*mongo.Client, error) {
	var err error
	var client *mongo.Client
	var connString connstring.ConnString

	if connString, err = connstring.Parse(uri); err != nil {
		return client, err
	}
	opt := options.Client().ApplyURI(uri)
	if connString.Username == "" {
		opt.Auth = nil
	}
	if len(opts) >= 2 && opts[0] != "" && opts[1] != "" {
		var caBytes []byte
		var clientBytes []byte
		if caBytes, err = ioutil.ReadFile(opts[0]); err != nil {
			return nil, err
		}
		if clientBytes, err = ioutil.ReadFile(opts[1]); err != nil {
			return nil, err
		}

		roots := x509.NewCertPool()
		if ok := roots.AppendCertsFromPEM(caBytes); !ok {
			panic("failed to parse root certificate")
		}
		certs, e := tls.X509KeyPair(clientBytes, clientBytes)
		if e != nil {
			log.Fatalf("invalid key pair: %v", e)
		}
		cfg := &tls.Config{RootCAs: roots, Certificates: []tls.Certificate{certs}}
		opt.SetTLSConfig(cfg)
	}
	if client, err = mongo.NewClient(opt); err != nil {
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

// ReadPasswordFromStdin reads password from stdin
func ReadPasswordFromStdin() (string, error) {
	// if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
	// 	return "", errors.New("Missing password")
	// }
	var buffer []byte
	var err error
	fmt.Print("Enter Password: ")
	if buffer, err = terminal.ReadPassword(int(syscall.Stdin)); err != nil {
		return "", err
	}
	return string(buffer), err
}
