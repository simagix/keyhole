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

// BSONPrinter stores bson printer info
type BSONPrinter struct {
	verbose bool
	version string
}

// NewBSONPrinter returns BSONPrinter
func NewBSONPrinter(version string) *BSONPrinter {
	return &BSONPrinter{version: version}
}

// SetVerbose sets verbose level
func (p *BSONPrinter) SetVerbose(verbose bool) {
	p.verbose = verbose
}

// Translate bson to json
func (p *BSONPrinter) Translate(filename string) error {
	var err error
	var str string
	var data []byte
	var doc bson.M
	var fd *bufio.Reader
	if fd, err = gox.NewFileReader(filename); err != nil {
		log.Fatal(err)
	}
	if data, err = ioutil.ReadAll(fd); err != nil {
		log.Fatal(err)
	}
	bson.Unmarshal(data, &doc)
	if p.verbose {
		fmt.Println(gox.Stringify(doc, "", "  "))
	}
	if doc["keyhole"] != nil {
		var logger Logger
		if buf, err := bson.Marshal(doc["keyhole"]); err == nil {
			bson.Unmarshal(buf, &logger)
			fmt.Println(logger.Print())
		} else {
			return err
		}
	} else {
		return errors.New("unsupported")
	}

	if strings.HasSuffix(filename, "-log.bson.gz") {
		li := NewLogInfo(p.version)
		if str, err = li.AnalyzeFile(filename, false); err != nil {
			return err
		}
		fmt.Println(str)
	} else if strings.HasSuffix(filename, "-index.bson.gz") {
		ix := NewIndexStats(p.version)
		if err = ix.SetIndexesMapFromFile(filename); err != nil {
			return err
		}
		ix.Print()
	} else if strings.HasSuffix(filename, "-stats.bson.gz") {
		ofile := filepath.Base(filename)
		ofile = (ofile)[:len(ofile)-7] + "json"
		if data, err = bson.MarshalExtJSON(doc, false, false); err == nil {
			ioutil.WriteFile(ofile, data, 0644)
			fmt.Println("JSON outputs to", ofile)
		} else {
			return err
		}
	} else {
		return errors.New("unsupported")
	}
	return err
}
