// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/simagix/gox"
	"go.mongodb.org/mongo-driver/bson"
)

// PrintBSON prints bson metadata
func PrintBSON(filename string) error {
	var err error
	var str string
	if strings.HasSuffix(filename, "-log.bson.gz") {
		li := NewLogInfo()
		if str, err = li.AnalyzeFile(filename, false); err != nil {
			return err
		}
		fmt.Println(str)
	} else if strings.HasSuffix(filename, "-index.bson.gz") {
		ix := NewIndexes(nil)
		if err = ix.SetIndexesMapFromFile(filename); err != nil {
			return err
		}
		ix.Print()
	} else if strings.HasSuffix(filename, ".bson.gz") {
		ofile := filepath.Base(filename)
		ofile = (ofile)[:len(ofile)-7] + "json"
		var doc bson.M
		var data []byte
		var err error
		var fd *bufio.Reader
		if fd, err = gox.NewFileReader(filename); err != nil {
			log.Fatal(err)
		}
		if data, err = ioutil.ReadAll(fd); err != nil {
			log.Fatal(err)
		}
		bson.Unmarshal(data, &doc)
		if data, err = bson.MarshalExtJSON(doc, false, false); err == nil {
			ioutil.WriteFile(ofile, data, 0644)
			fmt.Println("JSON outputs to", ofile)
		} else {
			return err
		}
	} else {
		return errors.New("unrecognized input")
	}
	return err
}
