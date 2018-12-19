// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"bufio"
	"compress/gzip"
	"os"
	"testing"
)

func TestNewReader(t *testing.T) {
	var err error
	var file *os.File
	filename := "/tmp/count.file"
	if file, err = os.Create(filename); err != nil {
		t.Fatal(err)
	}
	writer := gzip.NewWriter(file)
	str := "keyhole"
	b := []byte(str)
	writer.Write(b)
	writer.Flush()
	file.Close()

	file, _ = os.Open(filename)
	defer file.Close()
	reader, _ := NewReader(file)
	buf, _, _ := reader.ReadLine()

	if str != string(buf) {
		t.Fatal(string(buf))
	}
}

func TestCountLines(t *testing.T) {
	var err error
	var file *os.File
	filename := "/tmp/count.file"
	if file, err = os.Create(filename); err != nil {
		t.Fatal(err)
	}
	writer := bufio.NewWriter(file)
	total := 10

	for i := 0; i < total; i++ {
		writer.WriteString("value\n")
	}
	writer.Flush()
	file.Close()
	file, _ = os.Open(filename)
	defer file.Close()
	reader := bufio.NewReader(file)
	count, _ := CountLines(reader)

	if count != total {
		t.Fatal(count)
	}
}
