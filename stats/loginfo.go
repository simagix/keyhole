// Copyright 2018 Kuei-chun Chen. All rights reserved.

package stats

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/simagix/keyhole/utils"
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

// LogInfo -
func LogInfo(filename string) {
	var opsMap map[string]OpPerformanceDoc
	opsMap = make(map[string]OpPerformanceDoc)
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("error opening file ", err)
		return
	}
	defer file.Close()
	reader := utils.NewReader(file)
	lineCounts, _ := utils.CountLines(reader)

	matched := regexp.MustCompile(`^\S+ .? CONTROL\s+\[\w+\] (\w+(:)?) (.*)$`)
	file.Seek(0, 0)
	reader = utils.NewReader(file)
	for {
		buf, _, err := reader.ReadLine() // 0x0A separator = newline
		if err != nil {
			break
		} else if matched.MatchString(string(buf)) == true {
			result := matched.FindStringSubmatch(string(buf))
			if result[1] == "db" {
				fmt.Println("db", result[3])
			} else if result[1] == "options:" {
				re := regexp.MustCompile(`((\S+):)`)
				body := re.ReplaceAllString(result[3], "\"$1\":")
				var buf bytes.Buffer
				json.Indent(&buf, []byte(body), "", "  ")
				fmt.Println("config options:")
				fmt.Println(string(buf.Bytes()))
				break
			}
		}
	}

	matched = regexp.MustCompile(`^\S+ .? (\w+)\s+\[\w+\] (\w+) (\S+) \S+: (.*) (\d+)ms$`)
	file.Seek(0, 0)
	reader = utils.NewReader(file)
	index := 0

	for {
		if index%25 == 1 {
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
			} else if op == "count" {
				nstr := ""
				s := getDocByField(filter, "query: ")
				if s != "" {
					nstr = s
				} else {
					fmt.Println(op)
					fmt.Println(filter)
					os.Exit(0)
				}
				filter = nstr
			} else if op == "delete" || op == "update" || op == "remove" {
				s := getDocByField(filter, "q: ")
				if s != "" {
					filter = s
				}
			}

			// remove unneeded info
			// re = regexp.MustCompile(`(createIndexes: "\w+", |count: "\w+", |find: "\w+", |delete: "\w+", |update: "\w+", |, \$db: "\w+" |,? ?skip: \d+|, limit: \d+|, batchSize: \d+|, singleBatch: \w+)|, multi: \w+|, upsert: \w+|bypassDocumentValidation: \w+, |ordered: \w+, |, shardVersion: \[.*\]`)
			// filter = re.ReplaceAllString(filter, "")
			re = regexp.MustCompile(`(: "[^"]*"|: \d+|: new Date\(\d+?\)|: true|: false)`)
			filter = re.ReplaceAllString(filter, ":1")
			re = regexp.MustCompile(`, shardVersion: \[.*\]`)
			filter = re.ReplaceAllString(filter, "")
			re = regexp.MustCompile(`( ObjectId\('\S+'\))|( Timestamp\(\d+, \d+\))`)
			filter = re.ReplaceAllString(filter, "1")
			// re = regexp.MustCompile(`(q: {.*}), u: {.*}`)
			// filter = re.ReplaceAllString(filter, "$1")
			filter = removeInElements(filter)
			filter = strings.Replace(strings.Replace(filter, "{ ", "{", -1), " }", "}", -1)
			key := op + "." + filter
			_, ok := opsMap[key]
			milli, _ := strconv.Atoi(ms)
			if ok {
				max := opsMap[key].MaxMilli
				if milli > max {
					max = milli
				}
				x := opsMap[key].TotalMilli + milli
				y := opsMap[key].Count + 1
				opsMap[key] = OpPerformanceDoc{Command: opsMap[key].Command, Namespace: ns, Filter: opsMap[key].Filter, MaxMilli: max, TotalMilli: x, Count: y, Scan: scan}
			} else {
				opsMap[key] = OpPerformanceDoc{Command: op, Namespace: ns, Filter: filter, TotalMilli: milli, MaxMilli: milli, Count: 1, Scan: scan}
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

	fmt.Fprintf(os.Stderr, "\r100%% ")
	fmt.Println("\r+-------+--------+-------+-------+------+---------------------------------+-----------------------------------------------------------------------+")
	fmt.Printf("|Command|COLLSCAN| avg ms| max ms| Count| %-32s| %-70s|\n", "Namespace", "Query Pattern")
	fmt.Println("|-------+--------+-------+-------+------+---------------------------------+-----------------------------------------------------------------------|")
	for _, value := range arr {
		str := value.Filter
		if len(value.Command) > 13 {
			value.Command = value.Command[:13]
		}
		if len(value.Namespace) > 33 {
			value.Namespace = value.Namespace[:33]
		}
		if len(str) > 70 {
			// fmt.Println(value.Filter)
			str = value.Filter[:70]
			idx := strings.LastIndex(str, " ")
			str = value.Filter[:idx]
		}
		if value.Scan == COLLSCAN {
			fmt.Printf("|%-7s \x1b[31;1m%8s\x1b[0m %7.0f %7d %6d %-33s \x1b[31;1m%-71s\x1b[0m|\n", value.Command, value.Scan,
				float64(value.TotalMilli)/float64(value.Count), value.MaxMilli, value.Count, value.Namespace, str)
		} else {
			fmt.Printf("|%-7s \x1b[31;1m%8s\x1b[0m %7.0f %7d %6d %-33s %-71s|\n", value.Command, value.Scan,
				float64(value.TotalMilli)/float64(value.Count), value.MaxMilli, value.Count, value.Namespace, str)
		}
		if len(value.Filter) > 70 {
			remaining := value.Filter[len(str):]
			for i := 0; i < len(remaining); i += 70 {
				epos := i + 70
				var pstr string
				if epos > len(remaining) {
					epos = len(remaining)
					pstr = remaining[i:epos]
				} else {
					str = remaining[i:epos]
					idx := strings.LastIndex(str, " ")
					pstr = str[:idx]
					i -= (70 - idx)
				}
				if value.Scan == COLLSCAN {
					fmt.Printf("|%72s   \x1b[31;1m%-70s\x1b[0m||\n", " ", pstr)
				} else {
					fmt.Printf("|%72s   %-70s|\n", " ", pstr)
				}
			}
		}
	}
	fmt.Println("+-------+--------+-------+-------+------+---------------------------------+-----------------------------------------------------------------------+")
}

// convert $in: [...] to $in: [ ]
func removeInElements(str string) string {
	idx := strings.Index(str, "$in: [")
	if idx < 0 {
		return str
	}

	idx += 6
	cnt, epos := 0, 0
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

	str = str[:idx] + "..." + str[epos:]
	return str
}

var filters = []string{"count", "delete", "find", "remove", "update"}

func hasFilter(op string) bool {
	for _, f := range filters {
		if f == op {
			return true
		}
	}
	return false
}
