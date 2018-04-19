package stats

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// GetSession -
func GetSession(uri string, ssl bool, sslCA string) (*mgo.Session, error) {
	var session *mgo.Session
	var err error

	if ssl {
		roots := x509.NewCertPool()
		if ca, ferr := ioutil.ReadFile(sslCA); ferr == nil {
			roots.AppendCertsFromPEM(ca)
		}
		tlsConfig := &tls.Config{}
		tlsConfig.RootCAs = roots
		dialInfo, perr := mgo.ParseURL(uri)
		if perr != nil {
			panic(perr)
		}
		dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			tlsConfig := &tls.Config{}
			conn, derr := tls.Dial("tcp", addr.String(), tlsConfig)
			if derr != nil {
				panic(derr)
			}
			return conn, derr
		}
		session, err = mgo.DialWithInfo(dialInfo)
	} else {
		session, err = mgo.Dial(uri)
		if err != nil {
			panic(err)
		}
	}

	return session, err
}

// IsMaster - Execute isMaster
func IsMaster(session *mgo.Session) bson.M {
	session.SetMode(mgo.Monotonic, true)
	result := bson.M{}
	if err := session.DB("admin").Run(bson.D{{"isMaster", 1}}, &result); err != nil {
		fmt.Println(err)
	}
	return result
}
