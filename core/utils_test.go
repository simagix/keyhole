// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"os"
	"testing"

	"github.com/simagix/keyhole/mongo"
)

func TestParseDialInfo(t *testing.T) {
	uri := "mongodb://root:secret@keyhole-00-jgtm2.mongodb.net,keyhole-01-jgtm2.mongodb.net/test"
	dialInfo, err := mongo.ParseURL(uri)
	if err != nil {
		t.Fatal(err)
	}
	bytes, _ := json.MarshalIndent(dialInfo.Addrs, "", "  ")
	t.Log(string(bytes))
}

func TestParseDialInfoSrv(t *testing.T) {
	uri := "mongodb+srv://root:secret@keyhole-jgtm2.mongodb.net/test"
	dialInfo, err := mongo.ParseURL(uri)
	if err != nil {
		t.Fatal(err)
	}
	bytes, _ := json.MarshalIndent(dialInfo.Addrs, "", "  ")
	t.Log(string(bytes))
}

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
