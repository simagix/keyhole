// Copyright 2018 Kuei-chun Chen. All rights reserved.

package util

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// COLLSCAN constance
const COLLSCAN = "COLLSCAN"

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

func getDocByField(str string, field string) string {
	i := strings.Index(str, field)
	if i < 0 {
		return ""
	}
	str = strings.Trim(str[i+len(field):], " ")
	isFound := false
	bpos := 0 // begin position
	epos := 0 // end position
	for _, r := range str {
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
	return str[bpos:epos]
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

// LogInfo -
func LogInfo(filename string, collscan bool, silent ...bool) (string, error) {
	var err error
	var reader *bufio.Reader
	var file *os.File
	var opsMap map[string]OpPerformanceDoc

	opsMap = make(map[string]OpPerformanceDoc)
	if file, err = os.Open(filename); err != nil {
		return "", err
	}
	defer file.Close()

	if reader, err = NewReader(file); err != nil {
		return "", err
	}
	lineCounts, _ := CountLines(reader)

	file.Seek(0, 0)
	reader, _ = NewReader(file)
	var buffer bytes.Buffer
	if strs := getConfigOptions(reader); len(strs) > 0 {
		for _, s := range strs {
			buffer.WriteString(s + "\n")
		}
	}

	matched := regexp.MustCompile(`^\S+ .? (\w+)\s+\[\w+\] (\w+) (\S+) \S+: (.*) (\d+)ms$`)
	file.Seek(0, 0)
	if reader, err = NewReader(file); err != nil {
		return "", err
	}

	index := 0
	for {
		if index%25 == 1 && len(silent) == 0 {
			fmt.Fprintf(os.Stderr, "\r%3d%% ", (100*index)/lineCounts)
		}
		buf, _, err := reader.ReadLine() // 0x0A separator = newline
		index++
		scan := ""
		if err != nil {
			break
		} else if matched.MatchString(string(buf)) == true {
			str := string(buf)
			if strings.Index(str, "COLLSCAN") >= 0 {
				scan = COLLSCAN
			}
			if collscan == true && scan != COLLSCAN {
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
					nstr = nstr + ", sort: " + s
				}
				filter = nstr
			} else if op == "count" || op == "distinct" {
				nstr := ""
				s := getDocByField(filter, "query: ")
				if s != "" {
					nstr = s
				}
				filter = nstr
			} else if op == "delete" || op == "update" || op == "remove" {
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
			} else if op == "aggregate" {
				nstr := ""
				s := ""
				for _, mstr := range []string{"pipeline: [ { $match: ", "pipeline: [ { $sort: "} {
					s = getDocByField(result[4], mstr)
					if s != "" {
						nstr = s
						filter = nstr
						break
					}
				}
				if s == "" {
					continue
				}
			} else if op == "getMore" {
				nstr := ""
				s := getDocByField(result[4], "originatingCommand: ")

				if s != "" {
					for _, mstr := range []string{"filter: ", "pipeline: [ { $match: ", "pipeline: [ { $sort: "} {
						s = getDocByField(result[4], mstr)
						if s != "" {
							nstr = s
							filter = nstr
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
			if index == "" && strings.Index(str, "planSummary: IDHACK") >= 0 {
				index = "IDHACK"
			}
			if scan == "" && strings.Index(str, "planSummary: COUNT_SCAN") >= 0 {
				index = "COUNT_SCAN"
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
			re = regexp.MustCompile(`(: \/.*\/(.?) })`)
			filter = re.ReplaceAllString(filter, ": /regex/$2}")
			filter = strings.Replace(strings.Replace(filter, "{ ", "{", -1), " }", "}", -1)
			key := op + "." + filter + "." + scan
			_, ok := opsMap[key]
			milli, _ := strconv.Atoi(ms)
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

	arr := make([]OpPerformanceDoc, 0, len(opsMap))
	for _, value := range opsMap {
		arr = append(arr, value)
	}
	sort.Slice(arr, func(i, j int) bool {
		return float64(arr[i].TotalMilli)/float64(arr[i].Count) > float64(arr[j].TotalMilli)/float64(arr[j].Count)
	})
	if len(silent) == 0 {
		fmt.Fprintf(os.Stderr, "\r     \r")
	}
	return buffer.String() + printLogsSummary(arr), nil
}

func printLogsSummary(arr []OpPerformanceDoc) string {
	var buffer bytes.Buffer
	buffer.WriteString("\r+---------+--------+------+--------+------+---------------------------------+-------------------------------------------------------------+\n")
	buffer.WriteString(fmt.Sprintf("| Command |COLLSCAN|avg ms| max ms | Count| %-32s| %-60s|\n", "Namespace", "Query Pattern"))
	buffer.WriteString("|---------+--------+------+--------+------+---------------------------------+-------------------------------------------------------------|\n")
	for _, value := range arr {
		str := value.Filter
		if len(value.Command) > 13 {
			value.Command = value.Command[:13]
		}
		if len(value.Namespace) > 33 {
			length := len(value.Namespace)
			value.Namespace = value.Namespace[:1] + "*" + value.Namespace[(length-31):]
		}
		if len(str) > 70 {
			str = value.Filter[:70]
			idx := strings.LastIndex(str, " ")
			str = value.Filter[:idx]
		}
		output := ""
		avg := float64(value.TotalMilli) / float64(value.Count)
		avgstr := fmt.Sprintf("%6.0f", avg)
		if avg >= 3600000 {
			avg /= 3600000
			avgstr = fmt.Sprintf("%4.1fh", avg)
		} else if avg >= 60000 {
			avg /= 60000
			avgstr = fmt.Sprintf("%3.1fm", avg)
		} else if avg >= 1000 {
			avg /= 1000
			avgstr = fmt.Sprintf("%3.1fs", avg)
		}
		if value.Scan == COLLSCAN {
			output = fmt.Sprintf("|%-9s \x1b[31;1m%8s\x1b[0m %6s %8d %6d %-33s \x1b[31;1m%-61s\x1b[0m|\n", value.Command, value.Scan,
				avgstr, value.MaxMilli, value.Count, value.Namespace, str)
		} else {
			output = fmt.Sprintf("|%-9s \x1b[31;1m%8s\x1b[0m %6s %8d %6d %-33s %-61s|\n", value.Command, value.Scan,
				avgstr, value.MaxMilli, value.Count, value.Namespace, str)
		}
		buffer.WriteString(output)
		if len(value.Filter) > 60 {
			remaining := value.Filter[len(str):]
			for i := 0; i < len(remaining); i += 70 {
				epos := i + 60
				var pstr string
				if epos > len(remaining) {
					epos = len(remaining)
					pstr = remaining[i:epos]
				} else {
					str = strings.Trim(remaining[i:epos], " ")
					idx := strings.LastIndex(str, " ")
					if idx > 0 {
						pstr = str[:idx]
						i -= (60 - idx)
					}
				}
				if value.Scan == COLLSCAN {
					output = fmt.Sprintf("|%72s   \x1b[31;1m%-70s\x1b[0m||\n", " ", pstr)
					buffer.WriteString(output)
				} else {
					output = fmt.Sprintf("|%72s   %-70s|\n", " ", pstr)
					buffer.WriteString(output)
				}
				buffer.WriteString(output)
			}
		}
		if value.Index != "" {
			output = fmt.Sprintf("|---index: \x1b[32;1m%-71s\x1b[0m\n", value.Index)
			buffer.WriteString(output)
		}
	}
	buffer.WriteString("+---------+--------+------+--------+------+---------------------------------+-------------------------------------------------------------+\n")
	return buffer.String()
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

var filters = []string{"count", "delete", "find", "remove", "update", "aggregate", "getMore"}

func hasFilter(op string) bool {
	for _, f := range filters {
		if f == op {
			return true
		}
	}
	return false
}
