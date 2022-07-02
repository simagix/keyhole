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
	"golang.org/x/term"
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
		if !strings.Contains(host, ":") {
			host += ":27017"
		}
		dial := net.Dialer{Timeout: 2 * time.Second}
		var c net.Conn
		if c, err = dial.Dial("tcp", host); err != nil {
			return nil, err
		}
		c.Close()
	}
	opts := options.Client().ApplyURI(uri)
	if opts.AppName == nil {
		opts.SetAppName("Keyhole Lib")
	}
	if connString.Username == "" && connString.AuthMechanism == "" {
		opts.Auth = nil
	}
	if connString.SSLSet {
		tlsConfig := new(tls.Config)
		if connString.SSLInsecureSet {
			tlsConfig.InsecureSkipVerify = true
		}
		if connString.SSLCaFileSet {
			roots := x509.NewCertPool()
			var caBytes []byte
			if caBytes, err = ioutil.ReadFile(connString.SSLCaFile); err != nil {
				return nil, err
			}
			if ok := roots.AppendCertsFromPEM(caBytes); !ok {
				return client, errors.New("failed to parse root certificate")
			}
			tlsConfig.RootCAs = roots
		}
		var certs tls.Certificate
		if connString.SSLClientCertificateKeyFileSet {
			connString.SSL = true
			var clientBytes []byte
			if clientBytes, err = ioutil.ReadFile(connString.SSLClientCertificateKeyFile); err != nil {
				return nil, err
			}
			if certs, err = tls.X509KeyPair(clientBytes, clientBytes); err != nil {
				return nil, err
			}
			tlsConfig.Certificates = []tls.Certificate{certs}
			opts.SetTLSConfig(tlsConfig)
		}
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
	colon += begin
	at := strings.LastIndex(uri, "@")
	if colon > 0 && at > colon {
		portion := uri[colon+1 : at]
		if portion, err = url.QueryUnescape(portion); err != nil {
			return connString, err
		}
		uri = strings.Replace(uri, portion, url.QueryEscape(portion), 1)
	}
	connString, err = connstring.Parse(uri)                     // ignore error to accomodate authMechanism=PLAIN
	if connString.Username != "" && connString.Password == "" { // missing password, prompt for it
		fmt.Printf("Enter %v's Password: ", connString.Username)
		var data []byte
		if data, err = term.ReadPassword(int(syscall.Stdin)); err != nil {
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
