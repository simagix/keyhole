package utils

import (
	"bufio"
	"compress/gzip"
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
