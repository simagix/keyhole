// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"os"
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

// NewReader returns a reader from either a gzip or plain file
func NewReader(file *os.File) (*bufio.Reader, error) {
	var buf []byte
	var err error
	var reader *bufio.Reader

	reader = bufio.NewReader(file)
	if buf, err = reader.Peek(2); err != nil {
		return reader, err
	}
	file.Seek(0, 0)

	if buf[0] == 31 && buf[1] == 139 {
		var zreader *gzip.Reader
		if zreader, err = gzip.NewReader(file); err != nil {
			return reader, err
		}
		reader = bufio.NewReader(zreader)
	} else {
		reader = bufio.NewReader(file)
	}

	return reader, nil
}

// CountLines count number of '\n'
func CountLines(reader *bufio.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	lineSep := []byte{'\n'}
	lineCounts := 0
	for {
		c, err := reader.Read(buf)
		lineCounts += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return lineCounts, err

		case err != nil:
			return lineCounts, err
		}
	}
}
