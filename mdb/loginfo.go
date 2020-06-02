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

// LogStats log stats structure
type LogStats struct {
	filter string
	index  string
	milli  int
	ns     string
	op     string
	scan   string
}

const dollarCmd = "$cmd"

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
		if len(li.OpsPatterns) > 0 {
			li.OutputFilename = filepath.Base(filename)
			if strings.HasSuffix(li.OutputFilename, ".gz") {
				li.OutputFilename = li.OutputFilename[:len(li.OutputFilename)-3]
			}
			if strings.HasSuffix(li.OutputFilename, ".log") == false {
				li.OutputFilename += ".log"
			}
			li.OutputFilename += ".enc"
			var data bytes.Buffer
			enc := gob.NewEncoder(&data)
			if err = enc.Encode(li); err != nil {
				log.Println("encode error:", err)
			}
			ioutil.WriteFile(li.OutputFilename, data.Bytes(), 0644)
		}
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
	var stat LogStats
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
				logType = "logv2"
				if stat, err = li.ParseLogv2(str); err != nil {
					continue
				}
			} else if regexp.MustCompile("^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}.*").MatchString(str) == true {
				logType = "text"
				if stat, err = li.ParseLog(str); err != nil {
					continue
				}
			} else {
				return errors.New("unsupported format")
			}
		} else if logType == "text" {
			if stat, err = li.ParseLog(str); err != nil {
				continue
			}
		} else if logType == "logv2" {
			if stat, err = li.ParseLogv2(str); err != nil {
				continue
			}
		}
		if stat.op == "" {
			if li.verbose {
				fmt.Println(str)
			}
			continue
		} else if stat.op == dollarCmd {
			continue
		}
		key := stat.op + "." + stat.ns + "." + stat.filter + "." + stat.scan
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

// printLogsSummary prints loginfo summary
func (li *LogInfo) printLogsSummary() string {
	red := codeRed
	green := codeGreen
	tail := codeDefault
	if li.silent == true {
		red = ""
		green = ""
		tail = ""
	}
	summaries := []string{}
	if li.verbose == true {
		summaries = append([]string{}, li.mongoInfo)
	}
	if len(li.SlowOps) > 0 && li.verbose == true {
		summaries = append(summaries, fmt.Sprintf("Ops slower than 10 seconds (list top %d):", len(li.SlowOps)))
		for _, op := range li.SlowOps {
			summaries = append(summaries, fmt.Sprintf("%s (%s) %dms", op.Log, gox.MilliToTimeString(float64(op.Milli)), op.Milli))
		}
		summaries = append(summaries, "\n")
	}
	var buffer bytes.Buffer
	buffer.WriteString("\r+----------+--------+------+--------+------+---------------------------------+--------------------------------------------------------------+\n")
	buffer.WriteString(fmt.Sprintf("| Command  |COLLSCAN|avg ms| max ms | Count| %-32s| %-60s |\n", "Namespace", "Query Pattern"))
	buffer.WriteString("|----------+--------+------+--------+------+---------------------------------+--------------------------------------------------------------|\n")
	for _, value := range li.OpsPatterns {
		str := value.Filter
		if len(value.Command) > 10 {
			value.Command = value.Command[:10]
		}
		if len(value.Namespace) > 33 {
			length := len(value.Namespace)
			value.Namespace = value.Namespace[:1] + "*" + value.Namespace[(length-31):]
		}
		if len(str) > 60 {
			str = value.Filter[:60]
			idx := strings.LastIndex(str, " ")
			if idx > 0 {
				str = value.Filter[:idx]
			}
		}
		output := ""
		avg := float64(value.TotalMilli) / float64(value.Count)
		avgstr := gox.MilliToTimeString(avg)
		if value.Scan == COLLSCAN {
			output = fmt.Sprintf("|%-10s %v%8s%v %6s %8d %6d %-33s %v%-62s%v|\n", value.Command, red, value.Scan, tail,
				avgstr, value.MaxMilli, value.Count, value.Namespace, red, str, tail)
		} else {
			output = fmt.Sprintf("|%-10s %8s %6s %8d %6d %-33s %-62s|\n", value.Command, value.Scan,
				avgstr, value.MaxMilli, value.Count, value.Namespace, str)
		}
		buffer.WriteString(output)
		if len(value.Filter) > 60 {
			remaining := value.Filter[len(str):]
			for i := 0; i < len(remaining); i += 60 {
				epos := i + 60
				var pstr string
				if epos > len(remaining) {
					epos = len(remaining)
					pstr = remaining[i:epos]
				} else {
					str = strings.Trim(remaining[i:epos], " ")
					idx := strings.LastIndex(str, " ")
					if idx >= 0 {
						pstr = str[:idx]
						i -= (60 - idx)
					} else {
						pstr = str
						i -= (60 - len(str))
					}
				}
				if value.Scan == COLLSCAN {
					output = fmt.Sprintf("|%74s   %v%-62s%v|\n", " ", red, pstr, tail)
					buffer.WriteString(output)
				} else {
					output = fmt.Sprintf("|%74s   %-62s|\n", " ", pstr)
					buffer.WriteString(output)
				}
			}
		}
		if value.Index != "" {
			output = fmt.Sprintf("|...index:  %v%-128s%v|\n", green, value.Index, tail)
			buffer.WriteString(output)
		}
	}
	buffer.WriteString("+----------+--------+------+--------+------+---------------------------------+--------------------------------------------------------------+\n")
	summaries = append(summaries, buffer.String())
	return strings.Join(summaries, "\n")
}
