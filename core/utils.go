// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"syscall"

	"github.com/globalsign/mgo/bson"
	"golang.org/x/crypto/ssh/terminal"
)

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

// GetOptime -
func GetOptime(optime interface{}) int64 {
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
	return int64(ts) >> 32
}
