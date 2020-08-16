// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/simagix/gox"
)

// COLLSCAN constance
const COLLSCAN = "COLLSCAN"

// LogInfo keeps loginfo struct
type LogInfo struct {
	Collscan   bool        `bson:"collscan"`
	Logger     *Logger     `bson:"keyhole"`
	LogType    string      `bson:"type"`
	Regex      string      `bson:"regex"`
	OpPatterns []OpPattern `bson:"opPatterns"`
	Redaction  bool        `bson:"redact"`
	SlowOps    []SlowOps   `bson:"slowOps"`

	filename string
	regex    string
	silent   bool
	verbose  bool
}

// OpPattern stores performance data
type OpPattern struct {
	Command    string `bson:"command"`    // count, delete, find, remove, and update
	Count      int    `bson:"count"`      // number of ops
	Filter     string `bson:"filter"`     // query pattern
	MaxMilli   int    `bson:"maxmilli"`   // max millisecond
	Namespace  string `bson:"ns"`         // database.collectin
	Scan       string `bson:"scan"`       // COLLSCAN
	TotalMilli int    `bson:"totalmilli"` // total milliseconds
	Index      string `bson:"index"`      // index used
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
func NewLogInfo(version string) *LogInfo {
	li := LogInfo{Logger: NewLogger(version, "-loginfo"), Collscan: false, silent: false, verbose: false}
	li.regex = `^\S+ \S+\s+(\w+)\s+\[\w+\] (\w+) (\S+) \S+: (.*) (\d+)ms$` // SERVER-37743
	return &li
}

// SetCollscan -
func (li *LogInfo) SetCollscan(collscan bool) {
	li.Collscan = collscan
}

// SetRedaction sets redaction
func (li *LogInfo) SetRedaction(redaction bool) {
	li.Redaction = redaction
}

// SetSilent -
func (li *LogInfo) SetSilent(silent bool) {
	li.silent = silent
}

// SetVerbose -
func (li *LogInfo) SetVerbose(verbose bool) {
	li.verbose = verbose
}

// SetRegexPattern sets Regex patthen
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

// AnalyzeFile analyze logs from a file
func (li *LogInfo) AnalyzeFile(filename string) error {
	var err error
	li.filename = filename
	if strings.HasSuffix(filename, "-log.bson.gz") == true {
		var data []byte
		var err error
		var fd *bufio.Reader
		if fd, err = gox.NewFileReader(filename); err != nil {
			return err
		}
		if data, err = ioutil.ReadAll(fd); err != nil {
			return err
		}
		if err = bson.Unmarshal(data, &li); err != nil {
			return err
		}
	} else {
		li.LogType = ""
		var file *os.File
		var reader *bufio.Reader
		li.filename = filename
		if file, err = os.Open(filename); err != nil {
			return err
		}
		defer file.Close()
		if reader, err = gox.NewReader(file); err != nil {
			return err
		}
		var lineCounts int
		if li.silent == false {
			lineCounts, _ = gox.CountLines(reader)
		}
		file.Seek(0, 0)
		if reader, err = gox.NewReader(file); err != nil {
			return err
		}
		if err = li.Parse(reader, lineCounts); err != nil {
			return err
		}
	}
	return nil
}

// Parse parse text or json
func (li *LogInfo) Parse(reader *bufio.Reader, counts ...int) error {
	var err error
	var buf []byte
	var isPrefix bool
	var opsMap map[string]OpPattern
	var stat LogStats
	opsMap = make(map[string]OpPattern)
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
		if li.LogType == "" { //examine the log logType
			if regexp.MustCompile("^{.*}$").MatchString(str) == true {
				li.LogType = "logv2"
				if stat, err = li.ParseLogv2(str); err != nil {
					continue
				}
			} else if regexp.MustCompile("^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}.*").MatchString(str) == true {
				li.LogType = "text"
				if stat, err = li.ParseLog(str); err != nil {
					continue
				}
			} else {
				return errors.New("unsupported format")
			}
		} else if li.LogType == "text" {
			if stat, err = li.ParseLog(str); err != nil {
				continue
			}
		} else if li.LogType == "logv2" {
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
			opsMap[key] = OpPattern{Command: opsMap[key].Command, Namespace: stat.ns, Filter: opsMap[key].Filter,
				MaxMilli: max, TotalMilli: x, Count: y, Scan: stat.scan, Index: stat.index}
		} else {
			opsMap[key] = OpPattern{Command: stat.op, Namespace: stat.ns, Filter: stat.filter, TotalMilli: stat.milli,
				MaxMilli: stat.milli, Count: 1, Scan: stat.scan, Index: stat.index}
		}
	}
	li.OpPatterns = make([]OpPattern, 0, len(opsMap))
	for _, value := range opsMap {
		li.OpPatterns = append(li.OpPatterns, value)
	}
	sort.Slice(li.OpPatterns, func(i, j int) bool {
		return float64(li.OpPatterns[i].TotalMilli)/float64(li.OpPatterns[i].Count) > float64(li.OpPatterns[j].TotalMilli)/float64(li.OpPatterns[j].Count)
	})
	if li.silent == false {
		fmt.Fprintf(os.Stderr, "\r     \r")
	}
	return nil
}

// Save saves loginfo bson data
func (li *LogInfo) Save() error {
	var err error
	if len(li.OpPatterns) == 0 {
		return err
	}
	ofile := filepath.Base(li.filename)
	if strings.HasSuffix(ofile, ".gz") {
		ofile = ofile[:len(ofile)-3]
	}
	if strings.HasSuffix(ofile, ".log") == false {
		ofile += "-log.bson.gz"
	} else {
		ofile = ofile[:len(ofile)-4] + "-log.bson.gz"
	}
	if li.Redaction == true {
		li.SlowOps = []SlowOps{}
	}
	if li.LogType == "text" {
		li.Regex = li.regex
	} else if li.LogType == "logv2" {
		li.Regex = ""
	}
	var data []byte
	if data, err = bson.Marshal(li); err != nil {
		return err
	}
	outdir := "./out/"
	os.Mkdir(outdir, 0755)
	ofile = outdir + ofile
	gox.OutputGzipped(data, ofile)
	fmt.Println("bson log info written to", ofile)
	return nil
}

// Print prints indexes
func (li *LogInfo) Print() {
	fmt.Println(li.printLogsSummary())
	if li.verbose {
		var err error
		var data []byte
		if data, err = bson.MarshalExtJSON(li, false, false); err != nil {
			return
		}
		outdir := "./out/"
		os.Mkdir(outdir, 0755)
		ofile := outdir + strings.ReplaceAll(filepath.Base(li.filename), "bson.gz", "json")
		ioutil.WriteFile(ofile, data, 0644)
		fmt.Println("json data written to", ofile)
	}
}

// printLogsSummary prints loginfo summary
func (li *LogInfo) printLogsSummary() string {
	var maxSize = 10
	red := codeRed
	green := codeGreen
	tail := codeDefault
	if li.silent == true {
		red = ""
		green = ""
		tail = ""
	}
	summaries := []string{}
	var buffer bytes.Buffer
	buffer.WriteString("\r+----------+--------+------+--------+------+---------------------------------+--------------------------------------------------------------+\n")
	buffer.WriteString(fmt.Sprintf("| Command  |COLLSCAN|avg ms| max ms | Count| %-32s| %-60s |\n", "Namespace", "Query Pattern"))
	buffer.WriteString("|----------+--------+------+--------+------+---------------------------------+--------------------------------------------------------------|\n")
	count := 0
	for _, value := range li.OpPatterns {
		count++
		if count >= maxSize {
			break
		}
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
	summaries = append(summaries, fmt.Sprintf(`top %d of %v lines displayed; see HTML report for details.`, count, len(li.OpPatterns)))
	return strings.Join(summaries, "\n")
}
