// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"runtime"
	"strings"
	"syscall"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
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
			log.Println("Error:", pe)
			log.Println("dialInfo.Addrs:", dialInfo.Addrs)
			return nil, pe
		}
		if strings.Index(uri, "?") < 0 {
			uri = uri + "?" + params[0]
		} else {
			uri = uri + "&" + params[0]
		}

		dialInfo, err = mgo.ParseURL(uri)
		if err != nil {
			log.Println("Error:", err)
			return dialInfo, err
		}
		_, addrs, le := net.LookupSRV("mongodb", "tcp", srvAddr)
		if le != nil {
			log.Println("Error:", le)
			log.Println("dialInfo.Addrs:", dialInfo.Addrs)
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

// ReadPasswordFromStdin reads password from stdin
func ReadPasswordFromStdin() (string, error) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		return "", errors.New("Missing password")
	}
	var buffer []byte
	var err error
	fmt.Print("Enter Password: ")
	if buffer, err = terminal.ReadPassword(int(syscall.Stdin)); err != nil {
		return "", err
	}
	return string(buffer), err
}

// GetUint32 -
func GetUint32(r io.Reader) uint32 {
	var size uint32
	binary.Read(r, binary.LittleEndian, &size)
	return size
}

// GetOptime -
func GetOptime(optime interface{}) bson.MongoTimestamp {
	var ts bson.MongoTimestamp
	switch optime.(type) {
	case bson.M:
		bm := optime.(bson.M)
		b, _ := json.Marshal(bm)
		var optm OptimeDoc
		json.Unmarshal(b, &optm)
		ts = optm.TS
	case bson.MongoTimestamp:
		ts = optime.(bson.MongoTimestamp)
	}
	ts = ts >> 32
	return ts
}
