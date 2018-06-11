// Copyright 2018 Kuei-chun Chen. All rights reserved.

package utils

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"io"
	"os"
)

// NewReader returns a reader from either a gzip or plain file
func NewReader(file *os.File) *bufio.Reader {
	reader := bufio.NewReader(file)
	bytes, _ := reader.Peek(2)
	file.Seek(0, 0)

	if bytes[0] == 31 && bytes[1] == 139 {
		r, _ := gzip.NewReader(file)
		reader = bufio.NewReader(r)
	} else {
		reader = bufio.NewReader(file)
	}

	return reader
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
