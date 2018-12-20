// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mongo

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"strings"

	"github.com/globalsign/mgo"
)

// ParseURL supports seedlist connection string mongodb+srv://
func ParseURL(url string) (*mgo.DialInfo, error) {
	isSRV := false
	if strings.Index(url, "mongodb+srv://") == 0 {
		isSRV = true
		// *ssl = true
		url = "mongodb://" + (url)[14:]
		if strings.Index(url, "ssl=") < 0 {
			if strings.Index(url, "?") < 0 {
				url = url + "?ssl=true"
			} else {
				url = url + "&ssl=true"
			}
		}
	}

	dialInfo, err := mgo.ParseURL(url)
	if err != nil {
		return dialInfo, err
	}

	if isSRV == true {
		srvAddr := dialInfo.Addrs[0]
		params, pe := net.LookupTXT(srvAddr)
		if pe != nil {
			return nil, pe
		}
		if strings.Index(url, "?") < 0 {
			url = url + "?" + params[0]
		} else {
			url = url + "&" + params[0]
		}

		dialInfo, err = mgo.ParseURL(url)
		if err != nil {
			return dialInfo, err
		}
		_, addrs, le := net.LookupSRV("mongodb", "tcp", srvAddr)
		if le != nil {
			return nil, le
		}
		addresses := make([]string, len(addrs))
		for i, addr := range addrs {
			address := strings.TrimSuffix(addr.Target, ".")
			addresses[i] = fmt.Sprintf("%s:%d", address, addr.Port)
		}
		dialInfo.Addrs = addresses
	}

	return dialInfo, err
}

// AddCertificates includes certs in dialInfo
func AddCertificates(dialInfo *mgo.DialInfo, sslCA string, sslPEMKeyFile string) error {
	var err error
	if sslCA == "" && sslPEMKeyFile == "" {
		return err
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
		return errors.New("Missing PEM key file, include --PEMKeyFile flag")
	}

	if sslCA != "" {
		var roots *x509.CertPool
		if ca, ferr := ioutil.ReadFile(sslCA); ferr == nil {
			roots = x509.NewCertPool()
			roots.AppendCertsFromPEM(ca)
		}
		tlsConfig.RootCAs = roots
	} else {
		return errors.New("Missing CA file, include --CAFile flag")
	}

	dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
		return tls.Dial("tcp", addr.String(), tlsConfig)
	}
	return err
}
