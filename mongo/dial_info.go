// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mongo

import (
	"fmt"
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
