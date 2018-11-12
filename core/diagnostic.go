// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/binary"
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
}

// NewDiagnosticData -
func NewDiagnosticData() *DiagnosticData {
	return &DiagnosticData{ServerStatusList: []ServerStatusDoc{}, ReplSetStatusList: []ReplSetStatusDoc{}}
}

// PrintDiagnosticData prints diagnostic data of MongoD
func (d *DiagnosticData) PrintDiagnosticData(filename string, span int, isWeb bool) (string, error) {
	var err error
	var fi os.FileInfo

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
			log.Println(err)
			if err = d.ReadDiagnosticFile(filename); err != nil {
				return "", err
			}
		}
	}

	if d.ServerInfo != nil {
		b, _ := json.MarshalIndent(d.ServerInfo, "", "  ")
		log.Println(string(b))
	}

	if span < 0 {
		span = int(d.ServerStatusList[(len(d.ServerStatusList)-1)].LocalTime.Sub(d.ServerStatusList[0].LocalTime).Seconds()) / 20
	}

	return PrintAllStats(d.ServerStatusList, span), err
}

// ReadDiagnosticDir reads diagnotics.data from a directory
func (d *DiagnosticData) ReadDiagnosticDir(dirname string) error {
	var err error
	var files []os.FileInfo

	if files, err = ioutil.ReadDir(dirname); err != nil {
		return err
	}

	for _, f := range files {
		if strings.Index(f.Name(), "metrics.") != 0 && strings.Index(f.Name(), "keyhole_stats.") != 0 {
			continue
		}
		filename := dirname + "/" + f.Name()

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
	var docs ServerStatusDoc
	var repls ReplSetStatusDoc
	var metrics SystemMetricsDoc
	var pos int32

	if buffer, err = ioutil.ReadFile(filename); err != nil {
		return err
	}

	fmt.Print("reading ", filename)
	var doc = bson.M{}
	var r io.ReadCloser
	var cnt int
	for {
		if pos >= int32(len(buffer)) {
			break
		}
		bs := buffer[pos:(pos + 4)]
		var length int32

		if err = binary.Read(bytes.NewReader(bs), binary.LittleEndian, &length); err != nil {
			return err
		}

		bs = buffer[pos:(pos + length)]
		var out = bson.M{}
		err = bson.Unmarshal(bs, &out)
		pos += length

		if err != nil {
			continue
		} else if out["type"] == 0 {
			d.ServerInfo = out["doc"]
		} else if out["type"] == 1 {
			buf := out["data"].([]byte)
			// zlib decompress
			buf = buf[4:]
			bytesBuf := bytes.NewReader(buf)
			if r, err = zlib.NewReader(bytesBuf); err != nil {
				return err
			}
			var bytesBufWriter bytes.Buffer
			writer := bufio.NewWriter(&bytesBufWriter)
			io.Copy(writer, r)
			r.Close()
			bson.Unmarshal(bytesBufWriter.Bytes(), &doc)
			// systemMetrics
			// end
			// start
			// serverStatus
			// replSetGetStatus
			// local.oplog.rs.stats

			if doc["serverStatus"] != nil {
				cnt++
				buf, _ := json.Marshal(doc["serverStatus"])
				json.Unmarshal(buf, &docs)
				d.ServerStatusList = append(d.ServerStatusList, docs)
			}

			if doc["replSetGetStatus"] != nil {
				buf, _ := json.Marshal(doc["replSetGetStatus"])
				json.Unmarshal(buf, &repls)
				d.ReplSetStatusList = append(d.ReplSetStatusList, repls)
			}

			if doc["systemMetrics"] != nil {
				buf, _ := json.Marshal(doc["systemMetrics"])
				json.Unmarshal(buf, &metrics)
				d.SystemMetricsList = append(d.SystemMetricsList, metrics)
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
		return errors.New("Not doc found")
	}

	d.ServerStatusList = append(d.ServerStatusList, allDocs...)
	d.ReplSetStatusList = append(d.ReplSetStatusList, allRepls...)
	return err
}
