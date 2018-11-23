// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/globalsign/mgo/bson"
)

// DiagnosticData -
type DiagnosticData struct {
	ServerInfo        interface{}
	ServerStatusList  []ServerStatusDoc
	ReplSetStatusList []ReplSetStatusDoc
	SystemMetricsList []SystemMetricsDoc
	verbose           bool
}

// NewDiagnosticData -
func NewDiagnosticData(verbose bool) *DiagnosticData {
	return &DiagnosticData{ServerStatusList: []ServerStatusDoc{}, ReplSetStatusList: []ReplSetStatusDoc{}, verbose: verbose}
}

// PrintDiagnosticData prints diagnostic data of MongoD
func (d *DiagnosticData) PrintDiagnosticData(filenames []string, span int, isWeb bool) (string, error) {
	var err error
	var fi os.FileInfo

	for _, filename := range filenames {
		if fi, err = os.Stat(filename); err != nil {
			return "", err
		}
		switch mode := fi.Mode(); {
		case mode.IsDir():
			if err = d.ReadDiagnosticDir(filename); err != nil {
				return "", err
			}
		case mode.IsRegular():
			if err = d.analyzeServerStatus(filename); err != nil {
				if err = d.ReadDiagnosticFile(filename); err != nil {
					return "", err
				}
			}
		}
	}

	if d.ServerInfo != nil {
		b, _ := json.MarshalIndent(d.ServerInfo, "", "  ")
		log.Println(string(b))
	}

	if len(d.ServerStatusList) == 0 {
		return "No FTDC data found.", err
	}

	if isWeb == true {
		str := d.ServerStatusList[0].LocalTime.Format("2006-01-02T15:04:05Z") +
			" - " + d.ServerStatusList[len(d.ServerStatusList)-1].LocalTime.Format("2006-01-02T15:04:05Z")
		return str, err
	}
	return PrintAllStats(d.ServerStatusList, span), err
}

// ReadDiagnosticDir reads diagnotics.data from a directory
func (d *DiagnosticData) ReadDiagnosticDir(dirname string) error {
	var err error
	var files []os.FileInfo
	var filenames []string

	if files, err = ioutil.ReadDir(dirname); err != nil {
		return err
	}

	for _, f := range files {
		if strings.Index(f.Name(), "metrics.") != 0 && strings.Index(f.Name(), "keyhole_stats.") != 0 {
			continue
		}
		filename := dirname + "/" + f.Name()
		filenames = append(filenames, filename)
	}

	return d.ReadDiagnosticFiles(filenames)
}

// ReadDiagnosticFiles reads multiple files
func (d *DiagnosticData) ReadDiagnosticFiles(filenames []string) error {
	var err error

	for _, filename := range filenames {
		if err = d.analyzeServerStatus(filename); err != nil {
			if err = d.ReadDiagnosticFile(filename); err != nil {
				return err
			}
		}
	}
	return err
}

// ReadDiagnosticFile reads diagnostic.data from a file
func (d *DiagnosticData) ReadDiagnosticFile(filename string) error {
	btm := time.Now()
	var buffer []byte
	var err error
	var pos uint32
	if buffer, err = ioutil.ReadFile(filename); err != nil {
		return err
	}

	fmt.Print("reading ", filename)
	var r io.ReadCloser
	var cnt int

	for {
		if pos >= uint32(len(buffer)) {
			break
		}
		bs := buffer[pos:(pos + 4)]
		var length = GetUint32(bytes.NewReader(bs))

		bs = buffer[pos:(pos + length)]
		pos += length
		var out = bson.M{}
		if err = bson.Unmarshal(bs, &out); err != nil {
			log.Println(err)
			continue
		} else if out["type"] == 0 {
			d.ServerInfo = out["doc"]
		} else if out["type"] == 1 {
			bytesBuf := bytes.NewReader(out["data"].([]byte)[4:])
			// zlib decompress
			if r, err = zlib.NewReader(bytesBuf); err != nil {
				log.Println(err)
				continue
			}
			var data []byte
			if data, err = ioutil.ReadAll(r); err != nil {
				log.Println(err)
				continue
			}

			// systemMetrics
			// end
			// start
			// serverStatus
			// replSetGetStatus
			// local.oplog.rs.stats

			cnt++
			if d.verbose == true {
				if err = d.decodeFTDC(data); err != nil {
					return err
				}
			} else {
				d.decodeFirstDoc(data)
			}
		} else {
			log.Println("==>", out["type"])
		}
	}
	etm := time.Now()
	fmt.Println(", read:", cnt, ",time spent:", etm.Sub(btm).String())
	return err
}

// analyzeServerStatus -
func (d *DiagnosticData) analyzeServerStatus(filename string) error {
	var err error
	var file *os.File
	var reader *bufio.Reader
	var allDocs = []ServerStatusDoc{}
	var docs = []ServerStatusDoc{}
	var allRepls = []ReplSetStatusDoc{}
	var repls = []ReplSetStatusDoc{}

	if file, err = os.Open(filename); err != nil {
		return err
	}
	defer file.Close()

	if reader, err = NewReader(file); err != nil {
		return err
	}

	for {
		line, ferr := reader.ReadBytes('\n')
		if ferr == io.EOF {
			break
		}

		docs = []ServerStatusDoc{}
		if err = json.Unmarshal(line, &docs); err == nil {
			if len(docs) > 0 && docs[0].Host != "" {
				allDocs = append(allDocs, docs...)
			} else if err = json.Unmarshal(line, &repls); err == nil { // ReplSetStatusDoc
				allRepls = append(allRepls, repls...)
			}
		}
	}

	if len(allDocs) == 0 && len(allRepls) == 0 {
		return errors.New("No doc found")
	}

	d.ServerStatusList = append(d.ServerStatusList, allDocs...)
	d.ReplSetStatusList = append(d.ReplSetStatusList, allRepls...)
	return err
}

func (d *DiagnosticData) decodeFirstDoc(data []byte) {
	var doc = bson.M{}
	bson.Unmarshal(data, &doc) // first document
	var docs ServerStatusDoc
	if doc["serverStatus"] != nil {
		buf, _ := json.Marshal(doc["serverStatus"])
		json.Unmarshal(buf, &docs)
		d.ServerStatusList = append(d.ServerStatusList, docs)
	}

	var metrics SystemMetricsDoc
	if doc["systemMetrics"] != nil {
		buf, _ := json.Marshal(doc["systemMetrics"])
		json.Unmarshal(buf, &metrics)
		d.SystemMetricsList = append(d.SystemMetricsList, metrics)
	}

	var repls ReplSetStatusDoc
	if doc["replSetGetStatus"] != nil {
		buf, _ := json.Marshal(doc["replSetGetStatus"])
		json.Unmarshal(buf, &repls)
		d.ReplSetStatusList = append(d.ReplSetStatusList, repls)
	}
}
