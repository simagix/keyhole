// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/simagix/gox"
	"go.mongodb.org/mongo-driver/bson"
)

// BSONPrinter stores bson printer info
type BSONPrinter struct {
	version string
}

// NewBSONPrinter returns BSONPrinter
func NewBSONPrinter(version string) *BSONPrinter {
	return &BSONPrinter{version: version}
}

// Print print summary or output json from bson
func (p *BSONPrinter) Print(filename string) error {
	var err error
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
	if doc["keyhole"] == nil {
		return errors.New("unsupported")
	}
	var logger Logger
	if buf, err := bson.Marshal(doc["keyhole"]); err == nil {
		bson.Unmarshal(buf, &logger)
		fmt.Println(logger.Print())
	} else {
		return err
	}

	if strings.HasSuffix(filename, "-log.bson.gz") {
		li := NewLogInfo(p.version)
		if err = li.AnalyzeFile(filename); err != nil {
			return err
		}
		li.Print()
		if err = li.OutputJSON(); err != nil {
			return err
		}
	} else if strings.HasSuffix(filename, "-index.bson.gz") {
		ix := NewIndexStats(p.version)
		if err = ix.SetClusterDetailsFromFile(filename); err != nil {
			return err
		}
		ix.Print()
		ix.OutputJSON()
	} else if strings.HasSuffix(filename, ".bson.gz") {
		if strings.HasSuffix(filename, "-stats.bson.gz") {
			var cluster ClusterStats
			bson.Unmarshal(data, &cluster)
			fmt.Println(cluster.PrintShortSummary())
		}
		outdir := "./out/"
		os.Mkdir(outdir, 0755)
		ofile := filepath.Base(filename)
		idx := strings.Index(ofile, ".bson")
		ofile = outdir + (ofile)[:idx] + ".json"
		if data, err = bson.MarshalExtJSON(doc, false, false); err != nil {
			return err
		}
		ioutil.WriteFile(ofile, data, 0644)
		fmt.Println("json data written to", ofile)
	} else {
		return errors.New("unsupported")
	}
	return err
}
