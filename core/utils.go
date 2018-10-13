// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"fmt"
	"net"
	"runtime"
	"strings"
	"syscall"

	"github.com/globalsign/mgo"
	"golang.org/x/crypto/ssh/terminal"
)

// ParseDialInfo supports seedlist connection string mongodb+srv://
func ParseDialInfo(uri string) (*mgo.DialInfo, error) {
	isSRV := false
	if strings.Index(uri, "mongodb+srv://") == 0 {
		isSRV = true
		// *ssl = true
		uri = "mongodb://" + (uri)[14:]
		if strings.Index(uri, "ssl=") < 0 {
			if strings.Index(uri, "?") < 0 {
				uri = uri + "?ssl=true"
			} else {
				uri = uri + "&ssl=true"
			}
		}
	}

	dialInfo, err := mgo.ParseURL(uri)
	if err != nil {
		return dialInfo, err
	}

	if isSRV == true {
		srvAddr := dialInfo.Addrs[0]
		params, pe := net.LookupTXT(srvAddr)
		if pe != nil {
			fmt.Println("Error:", pe)
			fmt.Println("dialInfo.Addrs:", dialInfo.Addrs)
			return nil, pe
		}
		if strings.Index(uri, "?") < 0 {
			uri = uri + "?" + params[0]
		} else {
			uri = uri + "&" + params[0]
		}

		dialInfo, err = mgo.ParseURL(uri)
		if err != nil {
			fmt.Println("Error:", err)
			return dialInfo, err
		}
		_, addrs, le := net.LookupSRV("mongodb", "tcp", srvAddr)
		if le != nil {
			fmt.Println("Error:", le)
			fmt.Println("dialInfo.Addrs:", dialInfo.Addrs)
			return nil, le
		}
		addresses := make([]string, len(addrs))
		for i, addr := range addrs {
			address := strings.TrimSuffix(addr.Target, ".")
			addresses[i] = fmt.Sprintf("%s:%d", address, addr.Port)
		}
		dialInfo.Addrs = addresses
	}

	if dialInfo.Username != "" && dialInfo.Password == "" && (runtime.GOOS == "darwin" || runtime.GOOS == "linux") {
		fmt.Print("Enter Password: ")
		bytePassword, _ := terminal.ReadPassword(int(syscall.Stdin))
		dialInfo.Password = string(bytePassword)
	}
	return dialInfo, err
}
