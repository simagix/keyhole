// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/simagix/gox"
	"github.com/simagix/keyhole/sim/util"
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

// NewLogInfo -
func NewLogInfo(filename string) *LogInfo {
	li := LogInfo{filename: filename, collscan: false, silent: false, verbose: false}
	li.regex = `^\S+ \S+\s+(\w+)\s+\[\w+\] (\w+) (\S+) \S+: (.*) (\d+)ms$` // SERVER-37743
	li.OutputFilename = filepath.Base(filename)
	if strings.HasSuffix(li.OutputFilename, ".gz") {
		li.OutputFilename = li.OutputFilename[:len(li.OutputFilename)-3]
	}
	if strings.HasSuffix(li.OutputFilename, ".log") == false {
		li.OutputFilename += ".log"
	}
	li.OutputFilename += ".enc"
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

func getConfigOptions(reader *bufio.Reader) []string {
	matched := regexp.MustCompile(`^\S+ .? CONTROL\s+\[\w+\] (\w+(:)?) (.*)$`)
	var err error
	var buf []byte
	var strs []string

	for {
		buf, _, err = reader.ReadLine() // 0x0A separator = newline
		if err != nil {
			break
		} else if matched.MatchString(string(buf)) == true {
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

// Analyze -
func (li *LogInfo) Analyze() (string, error) {
	var err error

	if strings.HasSuffix(li.filename, ".enc") == true {
		var data []byte
		if data, err = ioutil.ReadFile(li.filename); err != nil {
			return "", err
		}
		buffer := bytes.NewBuffer(data)
		dec := gob.NewDecoder(buffer)
		if err = dec.Decode(li); err != nil {
			return "", err
		}
		li.OutputFilename = ""
	} else {
		if err = li.Parse(); err != nil {
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

// Parse -
func (li *LogInfo) Parse() error {
	var err error
	var reader *bufio.Reader
	var file *os.File
	var opsMap map[string]OpPerformanceDoc

	opsMap = make(map[string]OpPerformanceDoc)
	if file, err = os.Open(li.filename); err != nil {
		return err
	}
	defer file.Close()

	if reader, err = util.NewReader(file); err != nil {
		return err
	}
	lineCounts, _ := util.CountLines(reader)

	file.Seek(0, 0)
	reader, _ = util.NewReader(file)
	var buffer bytes.Buffer
	if strs := getConfigOptions(reader); len(strs) > 0 {
		for _, s := range strs {
			buffer.WriteString(s + "\n")
		}
	}
	li.mongoInfo = buffer.String()

	matched := regexp.MustCompile(li.regex)
	file.Seek(0, 0)
	if reader, err = util.NewReader(file); err != nil {
		return err
	}
	index := 0
	for {
		if index%25 == 1 && li.silent == false {
			fmt.Fprintf(os.Stderr, "\r%3d%% ", (100*index)/lineCounts)
		}
		var buf []byte
		var isPrefix bool
		buf, isPrefix, err = reader.ReadLine() // 0x0A separator = newline
		str := string(buf)
		for isPrefix == true {
			var bbuf []byte
			bbuf, isPrefix, err = reader.ReadLine()
			str += string(bbuf)
		}
		index++
		scan := ""
		aggStages := ""
		if err != nil {
			break
		} else if matched.MatchString(str) == true {
			if strings.Index(str, "COLLSCAN") >= 0 {
				scan = COLLSCAN
			}
			if li.collscan == true && scan != COLLSCAN {
				continue
			}
			result := matched.FindStringSubmatch(str)
			isFound := false
			bpos := 0 // begin position
			epos := 0 // end position
			for _, r := range result[4] {
				epos++
				if isFound == false && r == '{' {
					isFound = true
					bpos++
				} else if isFound == true {
					if r == '{' {
						bpos++
					} else if r == '}' {
						bpos--
					}
				}

				if isFound == true && bpos == 0 {
					break
				}
			}

			re := regexp.MustCompile(`^(\w+) ({.*})$`)
			op := result[2]
			ns := result[3]
			if ns == "local.oplog.rs" || strings.HasSuffix(ns, ".$cmd") == true {
				continue
			}
			filter := result[4][:epos]
			ms := result[5]
			if op == "command" {
				idx := strings.Index(filter, "command: ")
				if idx > 0 {
					filter = filter[idx+len("command: "):]
				}
				res := re.FindStringSubmatch(filter)
				if len(res) < 3 {
					continue
				}
				op = res[1]
				filter = res[2]
			}

			if hasFilter(op) == false {
				continue
			}
			if op == "delete" && strings.Index(filter, "writeConcern:") >= 0 {
				continue
			} else if op == "find" {
				nstr := "{ }"
				s := getDocByField(filter, "filter: ")
				if s != "" {
					nstr = s
				}
				s = getDocByField(filter, "sort: ")
				if s != "" {
					aggStages = ", sort: " + s
				}
				filter = nstr
			} else if op == "count" {
				nstr := ""
				s := getDocByField(filter, "query: ")
				if s != "" {
					nstr = s
				}
				filter = nstr
			} else if op == "distinct" {
				nstr := ""
				s := getDocByField(filter, "key: ")
				if s != "" {
					nstr = s
				}
				if strings.HasSuffix(nstr, " }") {
					nstr = nstr[:len(nstr)-2]
				}
				filter = "{" + nstr + ": 1}"
			} else if op == "delete" || op == "update" || op == "remove" || op == "findAndModify" {
				var s string
				// if result[1] == "WRITE" {
				if strings.Index(filter, "query: ") >= 0 {
					s = getDocByField(filter, "query: ")
				} else {
					s = getDocByField(filter, "q: ")
				}
				if s != "" {
					filter = s
				}
			} else if op == "aggregate" || (op == "getmore" && strings.Index(filter, "pipeline:") > 0) {
				s := ""
				for _, mstr := range []string{"pipeline: [ { $match: ", "pipeline: [ { $sort: "} {
					s = getDocByField(result[4], mstr)
					if s != "" {
						filter = s
						x := strings.Index(result[4], "$group: ")
						y := strings.Index(result[4], "$sort: ")
						if x > 0 && (x < y || y < 0) {
							aggStages = ", group: " + strings.ReplaceAll(getDocByField(result[4], "$group: "), "1.0", "1")
						}
						srt := getDocByField(result[4], "$sort: ")
						if srt != "" {
							aggStages += ", sort: " + strings.ReplaceAll(srt, "1.0", "1")
						}
						break
					}
				}
				if s == "" {
					if scan == "COLLSCAN" { // it's a collection scan without $match or $sort
						filter = "{}"
					} else {
						continue
					}
				}
			} else if op == "getMore" || op == "getmore" {
				s := getDocByField(result[4], "originatingCommand: ")
				if s != "" {
					s = getDocByField(s, "filter: ")
					for _, mstr := range []string{"filter: ", "pipeline: [ { $match: ", "pipeline: [ { $sort: "} {
						s = getDocByField(result[4], mstr)
						if s != "" {
							filter = s
							break
						}
					}
					if s == "" {
						continue
					}
				} else {
					continue
				}
			}
			index := getDocByField(str, "planSummary: IXSCAN")
			if index != "" {
				index = ""
				tmp := str
				i := strings.Index(tmp, " IXSCAN")
				for i >= 0 {
					index += getDocByField(tmp, " IXSCAN") + ",\n"
					tmp = tmp[i+7:]
					i = strings.Index(tmp, " IXSCAN")
				}
				index = index[:len(index)-2]
			} else if strings.Index(str, "planSummary: EOF") >= 0 {
				index = "EOF"
			} else if strings.Index(str, "planSummary: IDHACK") >= 0 {
				index = "IDHACK"
			} else if strings.Index(str, "planSummary: COUNT_SCAN") >= 0 {
				index = "COUNT_SCAN"
			} else if strings.Index(str, "planSummary: DISTINCT_SCAN") >= 0 {
				index = "DISTINCT_SCAN"
			}
			filter = removeInElements(filter, "$in: [ ")
			filter = removeInElements(filter, "$nin: [ ")
			filter = removeInElements(filter, "$in: [ ")
			filter = removeInElements(filter, "$nin: [ ")

			isRegex := strings.Index(filter, "{ $regex: ")
			if isRegex >= 0 {
				cnt := 0
				for _, r := range filter[isRegex:] {
					if r == '}' {
						break
					}
					cnt++
				}
				filter = filter[:(isRegex+10)] + "/.../.../" + filter[(isRegex+cnt):]
			}

			re = regexp.MustCompile(`(: "[^"]*"|: -?\d+(\.\d+)?|: new Date\(\d+?\)|: true|: false)`)
			filter = re.ReplaceAllString(filter, ":1")
			re = regexp.MustCompile(`, shardVersion: \[.*\]`)
			filter = re.ReplaceAllString(filter, "")
			re = regexp.MustCompile(`( ObjectId\('\S+'\))|(UUID\("\S+"\))|( Timestamp\(\d+, \d+\))|(BinData\(\d+, \S+\))`)
			filter = re.ReplaceAllString(filter, "1")
			re = regexp.MustCompile(`(: \/(\^)?\S+\/(\S+)? })`)
			filter = re.ReplaceAllString(filter, ": /${2}regex/$3}")
			filter = strings.Replace(strings.Replace(filter, "{ ", "{", -1), " }", "}", -1)
			// filter = reorderFilterFields(filter)
			filter += aggStages
			key := op + "." + filter + "." + scan
			_, ok := opsMap[key]
			milli, _ := strconv.Atoi(ms)
			if len(li.SlowOps) < 10 || milli > li.SlowOps[9].Milli {
				li.SlowOps = append(li.SlowOps, SlowOps{Milli: milli, Log: str})
				sort.Slice(li.SlowOps, func(i, j int) bool {
					return li.SlowOps[i].Milli > li.SlowOps[j].Milli
				})
				if len(li.SlowOps) > 10 {
					li.SlowOps = li.SlowOps[:10]
				}
			}

			if ok {
				max := opsMap[key].MaxMilli
				if milli > max {
					max = milli
				}
				x := opsMap[key].TotalMilli + milli
				y := opsMap[key].Count + 1
				opsMap[key] = OpPerformanceDoc{Command: opsMap[key].Command, Namespace: ns, Filter: opsMap[key].Filter, MaxMilli: max, TotalMilli: x, Count: y, Scan: scan, Index: index}
			} else {
				opsMap[key] = OpPerformanceDoc{Command: op, Namespace: ns, Filter: filter, TotalMilli: milli, MaxMilli: milli, Count: 1, Scan: scan, Index: index}
			}
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
	summaries := []string{}
	if li.verbose == true {
		summaries = append([]string{}, li.mongoInfo)
	}
	if len(li.SlowOps) > 0 && li.verbose == true {
		summaries = append(summaries, fmt.Sprintf("Ops slower than 10 seconds (list top %d):", len(li.SlowOps)))
		for _, op := range li.SlowOps {
			// summaries = append(summaries, MilliToTimeString(float64(op.Milli))+" => "+op.Log)
			summaries = append(summaries, fmt.Sprintf("%s (%s) %dms", op.Log, MilliToTimeString(float64(op.Milli)), op.Milli))
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
			str = value.Filter[:idx]
		}
		output := ""
		avg := float64(value.TotalMilli) / float64(value.Count)
		avgstr := MilliToTimeString(avg)
		if value.Scan == COLLSCAN {
			output = fmt.Sprintf("|%-10s \x1b[31;1m%8s\x1b[0m %6s %8d %6d %-33s \x1b[31;1m%-62s\x1b[0m|\n", value.Command, value.Scan,
				avgstr, value.MaxMilli, value.Count, value.Namespace, str)
		} else {
			output = fmt.Sprintf("|%-10s \x1b[31;1m%8s\x1b[0m %6s %8d %6d %-33s %-62s|\n", value.Command, value.Scan,
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
					output = fmt.Sprintf("|%74s   \x1b[31;1m%-62s\x1b[0m|\n", " ", pstr)
					buffer.WriteString(output)
				} else {
					output = fmt.Sprintf("|%74s   %-62s|\n", " ", pstr)
					buffer.WriteString(output)
				}
			}
		}
		if value.Index != "" {
			output = fmt.Sprintf("|...index:  \x1b[32;1m%-128s\x1b[0m|\n", value.Index)
			buffer.WriteString(output)
		}
	}
	buffer.WriteString("+----------+--------+------+--------+------+---------------------------------+--------------------------------------------------------------+\n")
	summaries = append(summaries, buffer.String())
	return strings.Join(summaries, "\n")
}

// convert $in: [...] to $in: [ ]
func removeInElements(str string, instr string) string {
	idx := strings.Index(str, instr)
	if idx < 0 {
		return str
	}

	idx += len(instr) - 1
	cnt, epos := -1, -1
	for _, r := range str {
		if cnt < idx {
			cnt++
			continue
		}
		if r == ']' {
			epos = cnt
			break
		}
		cnt++
	}

	if epos == -1 {
		str = str[:idx] + "...]"
	} else {
		str = str[:idx] + "..." + str[epos:]
	}
	return str
}

var filters = []string{"count", "delete", "find", "remove", "update", "aggregate", "getMore", "getmore", "findAndModify", "distinct"}

func hasFilter(op string) bool {
	for _, f := range filters {
		if f == op {
			return true
		}
	}
	return false
}

// MilliToTimeString converts milliseconds to time string, e.g. 1.5m
func MilliToTimeString(milli float64) string {
	avgstr := fmt.Sprintf("%6.0f", milli)
	if milli >= 3600000 {
		milli /= 3600000
		avgstr = fmt.Sprintf("%4.1fh", milli)
	} else if milli >= 60000 {
		milli /= 60000
		avgstr = fmt.Sprintf("%3.1fm", milli)
	} else if milli >= 1000 {
		milli /= 1000
		avgstr = fmt.Sprintf("%3.1fs", milli)
	}
	return avgstr
}

func getDocByField(str string, key string) string {
	ml := gox.NewMongoLog(str)
	return ml.Get(key)
}

func reorderFilterFields(str string) string {
	if strings.Index(str, "$and:") > 0 { // not able to parse it yet
		return str
	}
	filter := str[1 : len(str)-1]
	filter = strings.ReplaceAll(filter, ":", ": ")
	fields := strings.Fields(filter)
	mlog := gox.NewMongoLog(filter)
	m := map[string]string{}
	for _, field := range fields {
		if strings.HasSuffix(field, ":") == false {
			continue
		}
		field = field[:len(field)-1]
		if len(field) < 1 {
			continue
		}
		if field[0] == ' ' {
			field = field[1:]
		}
		if field[0] == '{' {
			field = field[1:]
		}
		if field[0] == '$' || field[0] == '/' {
			continue
		}
		if v := mlog.Get(field + ":"); v != "" {
			if strings.Index(field, " ") >= 0 {
				continue
			}
			m[field] = v
		}
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	filter = "{"
	for i, k := range keys {
		if i > 0 {
			filter += ", "
		}
		filter += k + ": " + strings.TrimSpace(m[k])
	}
	filter += "}"
	return filter
}
