// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"strings"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
	"golang.org/x/crypto/ssh/terminal"
)

// KeyholeDB default database
var KeyholeDB = "_KEYHOLE_88800"

// ExamplesCollection default test colection
var ExamplesCollection = "__examples"

// NewMongoClient new mongo client
func NewMongoClient(uri string) (*mongo.Client, error) {
	var err error
	var client *mongo.Client
	var connString connstring.ConnString
	if connString, err = connstring.Parse(uri); err != nil {
		return client, err
	}
	for _, host := range connString.Hosts {
		if strings.Index(host, ":") < 0 {
			host += ":27017"
		}
		if _, err = net.Dial("tcp", host); err != nil {
			return nil, err
		}
	}
	opts := options.Client().ApplyURI(uri)
	if opts.AppName == nil {
		opts.SetAppName("Keyhole Lib")
	}
	if connString.ReplicaSet == "" && len(connString.Hosts) == 1 && strings.HasPrefix(uri, "mongodb://") {
		opts.SetDirect(true)
	}
	if connString.Username == "" && connString.AuthMechanism == "" {
		opts.Auth = nil
	}
	if connString.SSLCaFile != "" {
		connString.SSL = true
		roots := x509.NewCertPool()
		var caBytes []byte
		if caBytes, err = ioutil.ReadFile(connString.SSLCaFile); err != nil {
			return nil, err
		}
		if ok := roots.AppendCertsFromPEM(caBytes); !ok {
			return client, errors.New("failed to parse root certificate")
		}
		var certs tls.Certificate
		if connString.SSLClientCertificateKeyFile != "" {
			var clientBytes []byte
			if clientBytes, err = ioutil.ReadFile(connString.SSLClientCertificateKeyFile); err != nil {
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
	return client, client.Connect(ctx)
}

// ParseURI checks if password is included
func ParseURI(uri string) (connstring.ConnString, error) {
	var err error
	var connString connstring.ConnString
	begin := strings.Index(uri, "://")
	begin += 3
	colon := strings.Index(uri[begin:], ":")
	at := strings.LastIndex(uri, "@")
	if colon > 0 && at > colon {
		colon += begin
		uri = strings.Replace(uri, uri[colon+1:at], url.QueryEscape(uri[colon+1:at]), 1)
	}
	connString, err = connstring.Parse(uri)                     // ignore error to accomodate authMechanism=PLAIN
	if connString.Username != "" && connString.Password == "" { // missing password, prompt for it
		fmt.Printf("Enter %v's Password: ", connString.Username)
		var data []byte
		if data, err = terminal.ReadPassword(int(syscall.Stdin)); err != nil {
			return connString, err
		}
		fmt.Println("")
		connString.Password = string(data)
		i := strings.Index(uri, connString.Username) + len(connString.Username)
		uri = (uri)[:i] + ":" + url.QueryEscape(connString.Password) + (uri)[i:]
		return connstring.Parse(uri)
	}
	return connString, err
}
