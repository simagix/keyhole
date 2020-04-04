// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/simagix/gox"
)

// COLLSCAN constance
const COLLSCAN = "COLLSCAN"

// LogInfo keeps loginfo struct
type LogInfo struct {
	OpsPatterns    []OpPerformanceDoc
	OutputFilename string
	SlowOps        []SlowOps
	collscan       bool
	filename       string
	mongoInfo      string
	regex          string
	silent         bool
	verbose        bool
}

// OpPerformanceDoc stores performance data
type OpPerformanceDoc struct {
	Command    string // count, delete, find, remove, and update
	Count      int    // number of ops
	Filter     string // query pattern
	MaxMilli   int    // max millisecond
	Namespace  string // database.collectin
	Scan       string // COLLSCAN
	TotalMilli int    // total milliseconds
	Index      string // index used
}

// SlowOps holds slow ops log and time
type SlowOps struct {
	Milli int
	Log   string
}

type logStats struct {
	filter string
	index  string
	milli  int
	ns     string
	op     string
	scan   string
}

// NewLogInfo -
func NewLogInfo() *LogInfo {
	li := LogInfo{collscan: false, silent: false, verbose: false}
	li.regex = `^\S+ \S+\s+(\w+)\s+\[\w+\] (\w+) (\S+) \S+: (.*) (\d+)ms$` // SERVER-37743
	return &li
}

// SetCollscan -
func (li *LogInfo) SetCollscan(collscan bool) {
	li.collscan = collscan
}

// SetSilent -
func (li *LogInfo) SetSilent(silent bool) {
	li.silent = silent
}

// SetVerbose -
func (li *LogInfo) SetVerbose(verbose bool) {
	li.verbose = verbose
}

// SetRegexPattern sets regex patthen
func (li *LogInfo) SetRegexPattern(regex string) {
	if regex != "" {
		li.regex = regex
	}
}

func getConfigOptions(buffers []string) []string {
	matched := regexp.MustCompile(`^\S+ .? CONTROL\s+\[\w+\] (\w+(:)?) (.*)$`)
	var strs []string

	for _, buf := range buffers {
		if matched.MatchString(buf) == true {
			result := matched.FindStringSubmatch(string(buf))
			if result[1] == "db" {
				s := "db " + result[3]
				strs = append(strs, s)
			} else if result[1] == "options:" {
				re := regexp.MustCompile(`((\S+):)`)
				body := re.ReplaceAllString(result[3], "\"$1\":")
				var buf bytes.Buffer
				json.Indent(&buf, []byte(body), "", "  ")

				strs = append(strs, "config options:")
				strs = append(strs, string(buf.Bytes()))
				return strs
			}
		}
	}
	return strs
}

const topN = 25

// Analyze -
func (li *LogInfo) Analyze(filename string) (string, error) {
	var err error

	if strings.HasSuffix(filename, ".enc") == true {
		var data []byte
		if data, err = ioutil.ReadFile(filename); err != nil {
			return "", err
		}
		buffer := bytes.NewBuffer(data)
		dec := gob.NewDecoder(buffer)
		if err = dec.Decode(li); err != nil {
			return "", err
		}
	} else {
		var file *os.File
		var reader *bufio.Reader
		li.filename = filename
		li.OutputFilename = filepath.Base(filename)
		if strings.HasSuffix(li.OutputFilename, ".gz") {
			li.OutputFilename = li.OutputFilename[:len(li.OutputFilename)-3]
		}
		if strings.HasSuffix(li.OutputFilename, ".log") == false {
			li.OutputFilename += ".log"
		}
		li.OutputFilename += ".enc"
		if file, err = os.Open(filename); err != nil {
			return "", err
		}
		defer file.Close()
		if reader, err = gox.NewReader(file); err != nil {
			return "", err
		}
		lineCounts, _ := gox.CountLines(reader)
		file.Seek(0, 0)
		if reader, err = gox.NewReader(file); err != nil {
			return "", err
		}
		if err = li.Parse(reader, lineCounts); err != nil {
			return "", err
		}
		var data bytes.Buffer
		enc := gob.NewEncoder(&data)
		if err = enc.Encode(li); err != nil {
			log.Println("encode error:", err)
		}
		ioutil.WriteFile(li.OutputFilename, data.Bytes(), 0644)
	}
	return li.printLogsSummary(), nil
}

// Parse parse text or json
func (li *LogInfo) Parse(reader *bufio.Reader, counts ...int) error {
	var err error
	var buf []byte
	var isPrefix bool
	var logType string
	var opsMap map[string]OpPerformanceDoc
	var stat logStats
	opsMap = make(map[string]OpPerformanceDoc)
	lineCounts := 0
	if len(counts) > 0 {
		lineCounts = counts[0]
	}
	index := 0
	for {
		if lineCounts > 0 && li.silent == false && index%50 == 0 {
			fmt.Fprintf(os.Stderr, "\r%3d%% ", (100*index)/lineCounts)
		}
		if buf, isPrefix, err = reader.ReadLine(); err != nil { // 0x0A separator = newline
			break
		}
		index++
		str := string(buf)
		for isPrefix == true {
			var bbuf []byte
			if bbuf, isPrefix, err = reader.ReadLine(); err != nil {
				break
			}
			str += string(bbuf)
		}
		if logType == "" { //examine the log logType
			if regexp.MustCompile("^{.*}$").MatchString(str) == true {
				logType = "json"
				if stat, err = li.parseJSON(str); err != nil {
					continue
				}
			} else if regexp.MustCompile("^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}.*").MatchString(str) == true {
				logType = "text"
				if stat, err = li.parseText(str); err != nil {
					continue
				}
			} else {
				return errors.New("unsupported format")
			}
		} else if logType == "text" {
			if stat, err = li.parseText(str); err != nil {
				continue
			}
		} else if logType == "json" {
			if stat, err = li.parseJSON(str); err != nil {
				continue
			}
		}

		key := stat.op + "." + stat.filter + "." + stat.scan
		_, ok := opsMap[key]
		if stat.op != "insert" && (len(li.SlowOps) < topN || stat.milli > li.SlowOps[topN-1].Milli) {
			li.SlowOps = append(li.SlowOps, SlowOps{Milli: stat.milli, Log: str})
			sort.Slice(li.SlowOps, func(i, j int) bool {
				return li.SlowOps[i].Milli > li.SlowOps[j].Milli
			})
			if len(li.SlowOps) > topN {
				li.SlowOps = li.SlowOps[:topN]
			}
		}

		if ok {
			max := opsMap[key].MaxMilli
			if stat.milli > max {
				max = stat.milli
			}
			x := opsMap[key].TotalMilli + stat.milli
			y := opsMap[key].Count + 1
			opsMap[key] = OpPerformanceDoc{Command: opsMap[key].Command, Namespace: stat.ns, Filter: opsMap[key].Filter,
				MaxMilli: max, TotalMilli: x, Count: y, Scan: stat.scan, Index: stat.index}
		} else {
			opsMap[key] = OpPerformanceDoc{Command: stat.op, Namespace: stat.ns, Filter: stat.filter, TotalMilli: stat.milli,
				MaxMilli: stat.milli, Count: 1, Scan: stat.scan, Index: stat.index}
		}
	}
	li.OpsPatterns = make([]OpPerformanceDoc, 0, len(opsMap))
	for _, value := range opsMap {
		li.OpsPatterns = append(li.OpsPatterns, value)
	}
	sort.Slice(li.OpsPatterns, func(i, j int) bool {
		return float64(li.OpsPatterns[i].TotalMilli)/float64(li.OpsPatterns[i].Count) > float64(li.OpsPatterns[j].TotalMilli)/float64(li.OpsPatterns[j].Count)
	})
	if li.silent == false {
		fmt.Fprintf(os.Stderr, "\r     \r")
	}
	return nil
}
