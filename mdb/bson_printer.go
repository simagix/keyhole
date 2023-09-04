// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/simagix/gox"
	"go.mongodb.org/mongo-driver/bson"
)

// BSONPrinter stores bson printer info
type BSONPrinter struct {
	nocolor bool
	version string
}

// NewBSONPrinter returns BSONPrinter
func NewBSONPrinter(version string) *BSONPrinter {
	return &BSONPrinter{version: version}
}

// SetNoColor set nocolor flag
func (p *BSONPrinter) SetNoColor(nocolor bool) {
	p.nocolor = nocolor
}

// Print print summary or output json from bson
func (p *BSONPrinter) Print(filename string) error {
	var err error
	var data []byte
	var doc bson.M
	var fd *bufio.Reader
	if fd, err = gox.NewFileReader(filename); err != nil {
		return err
	}
	if data, err = io.ReadAll(fd); err != nil {
		return err
	}
	bson.Unmarshal(data, &doc)
	if doc["keyhole"] == nil {
		return errors.New("unsupported, keyhole signature not found")
	}
	var logger gox.Logger
	if buf, err := bson.Marshal(doc["keyhole"]); err == nil {
		if err = bson.Unmarshal(buf, &logger); err != nil {
			return err
		}
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
		if err = ix.OutputJSON(); err != nil {
			return err
		}
	} else if strings.HasSuffix(filename, ".bson.gz") {
		if strings.HasSuffix(filename, "-perf.bson.gz") {
			type Perf struct {
				Logger  *gox.Logger         `bson:"keyhole"`
				Metrics map[string][]bson.M `bson:"metrics"`
				Results []string            `bson:"results"`
			}
			var perf Perf
			if err = bson.Unmarshal(data, &perf); err != nil {
				return err
			}
			for _, res := range perf.Results {
				fmt.Println(res)
			}
			delete(doc, "results")
		} else if strings.HasSuffix(filename, "-stats.bson.gz") {
			var cluster ClusterStats
			if err = bson.Unmarshal(data, &cluster); err != nil {
				return err
			}
			cluster.Print()
		}
		os.Mkdir(outdir, 0755)
		ofile := filepath.Base(filename)
		idx := strings.Index(ofile, ".bson")
		ofile = fmt.Sprintf(`%v/%v.json`, outdir, (ofile)[:idx])
		if data, err = bson.MarshalExtJSON(doc, false, false); err != nil {
			return err
		}
		if err = os.WriteFile(ofile, data, 0644); err != nil {
			return err
		}
		fmt.Println("json data written to", ofile)
	} else {
		return errors.New("unsupported")
	}
	return err
}
