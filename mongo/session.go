// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mongo

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net"

	"github.com/globalsign/mgo"
)

// GetSession connects to a mongod server without timeout timeout may interrupt load tests when
// having multiple go routines
func GetSession(dialInfo *mgo.DialInfo, sslCA string, sslPEMKeyFile string) (*mgo.Session, error) {
	if sslCA == "" && sslPEMKeyFile == "" {
		return mgo.DialWithInfo(dialInfo)
	}

	tlsConfig := &tls.Config{}
	tlsConfig.InsecureSkipVerify = true
	if sslPEMKeyFile != "" {
		clientCertPEM, _ := ioutil.ReadFile(sslPEMKeyFile)
		clientKeyPEM, _ := ioutil.ReadFile(sslPEMKeyFile)
		clientCert, _ := tls.X509KeyPair(clientCertPEM, clientKeyPEM)
		clientCert.Leaf, _ = x509.ParseCertificate(clientCert.Certificate[0])
		tlsConfig.Certificates = []tls.Certificate{clientCert}
	} else {
		return nil, errors.New("Missing PEM key file, include --PEMKeyFile flag")
	}

	if sslCA != "" {
		var roots *x509.CertPool
		if ca, ferr := ioutil.ReadFile(sslCA); ferr == nil {
			roots = x509.NewCertPool()
			roots.AppendCertsFromPEM(ca)
		}
		tlsConfig.RootCAs = roots
	} else {
		return nil, errors.New("Missing CA file, include --CAFile flag")
	}

	dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
		return tls.Dial("tcp", addr.String(), tlsConfig)
	}
	return mgo.DialWithInfo(dialInfo)
}
