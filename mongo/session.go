// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mongo

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"

	"github.com/globalsign/mgo"
)

// GetSession connects to a mongod server without timeout timeout may interrupt load tests when
// having multiple go routines
func GetSession(dialInfo *mgo.DialInfo, wmajor bool, ssl bool, sslCA string, sslPEMKeyFile string) (*mgo.Session, error) {
	if ssl {
		tlsConfig := &tls.Config{}
		tlsConfig.InsecureSkipVerify = true
		if sslPEMKeyFile != "" {
			clientCertPEM, _ := ioutil.ReadFile(sslPEMKeyFile)
			clientKeyPEM, _ := ioutil.ReadFile(sslPEMKeyFile)
			clientCert, _ := tls.X509KeyPair(clientCertPEM, clientKeyPEM)
			clientCert.Leaf, _ = x509.ParseCertificate(clientCert.Certificate[0])
			tlsConfig.Certificates = []tls.Certificate{clientCert}
		}

		if sslCA != "" {
			var roots *x509.CertPool
			if ca, ferr := ioutil.ReadFile(sslCA); ferr == nil {
				roots = x509.NewCertPool()
				roots.AppendCertsFromPEM(ca)
			}
			tlsConfig.RootCAs = roots
		}

		dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			return tls.Dial("tcp", addr.String(), tlsConfig)
		}
	}
	var session *mgo.Session
	var err error
	session, err = mgo.DialWithInfo(dialInfo)
	if err == nil {
		session.SetMode(mgo.Primary, true)
		if wmajor {
			session.SetSafe(&mgo.Safe{WMode: "majority"})
		}
	}
	return session, err
}
