// Copyright 2018 Kuei-chun Chen. All rights reserved.

package stats

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// OpPerformanceDoc stores performance data
type OpPerformanceDoc struct {
	Command   string // count, delete, find, remove, and update
	Namespace string // database.collectin
	Filter    string // query pattern
	Milli     int    // millisecond
	Count     int    // number of ops
	Scan      string // COLLSCAN
}

// LogInfo -
func LogInfo(filename string) {
	var opsMap map[string]OpPerformanceDoc
	opsMap = make(map[string]OpPerformanceDoc)
	var matched = regexp.MustCompile(`^\S+ .? (\w+)\s+\[\w+\] (\w+) (\S+) \S+: (.*) (\d+)ms$`)
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("error opening file ", err)
		return
	}
	defer file.Close()
	lineCounts, _ := countLines(file)
	file.Seek(0, 0)
	reader := bufio.NewReader(file)
	index := 0

	for {
		fmt.Printf("\r%3d%%", (100*index)/lineCounts)
		buf, _, err := reader.ReadLine() // 0x0A separator = newline
		index++
		scan := ""
		if err != nil {
			fmt.Printf("\r%3d%%\n", 100)
			break
		} else if matched.MatchString(string(buf)) == true {
			str := string(buf)
			if strings.Index(str, "COLLSCAN") >= 0 {
				scan = "COLLSCAN"
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

			// remove unneeded info
			re = regexp.MustCompile(`(createIndexes: "\w+", |count: "\w+", |find: "\w+", |delete: "\w+", |update: "\w+", |, \$db: "\w+" |,? ?skip: \d+|, limit: \d+|, batchSize: \d+|, singleBatch: \w+)|, multi: \w+|, upsert: \w+|, ordered: \w+|, shardVersion: \[ Timestamp 0\|0, ObjectId\('\S+'\) \]`)
			filter = re.ReplaceAllString(filter, "")
			re = regexp.MustCompile(`(: "[^"]*"|: \d+|: new Date\(\d+?\)|: true|: false)`)
			filter = re.ReplaceAllString(filter, ": 1")
			re = regexp.MustCompile(`(ObjectId\('\S+'\))`)
			filter = re.ReplaceAllString(filter, "1")
			re = regexp.MustCompile(`(q: {.*}), u: {.*}`)
			filter = re.ReplaceAllString(filter, "$1")
			filter = removeInElements(filter)
			key := op + "." + filter
			_, ok := opsMap[key]
			milli, _ := strconv.Atoi(ms)
			if ok {
				x := opsMap[key].Milli + milli
				y := opsMap[key].Count + 1
				opsMap[key] = OpPerformanceDoc{Command: opsMap[key].Command, Namespace: ns, Filter: opsMap[key].Filter, Milli: x, Count: y, Scan: scan}
			} else {
				opsMap[key] = OpPerformanceDoc{Command: op, Namespace: ns, Filter: filter, Milli: milli, Count: 1, Scan: scan}
			}
		}
	}

	arr := make([]OpPerformanceDoc, 0, len(opsMap))
	for _, value := range opsMap {
		arr = append(arr, value)
	}
	sort.Slice(arr, func(i, j int) bool {
		return float64(arr[i].Milli)/float64(arr[i].Count) > float64(arr[j].Milli)/float64(arr[j].Count)
	})
	fmt.Println("+-------+--------+-------+------+---------------------------------+----------------------------------------------------------------------+")
	fmt.Printf("|Command|COLLSCAN| avg ms| Count| %-32s| %-69s|\n", "Namespace", "Query Pattern")
	fmt.Println("|-------+--------+-------+------+---------------------------------+----------------------------------------------------------------------|")
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
		fmt.Printf("|%-7s|%8s|%7.0f|%6d|%-33s|%-70s|\n", value.Command, value.Scan, float64(value.Milli)/float64(value.Count), value.Count, value.Namespace, str)
		if len(value.Filter) > 70 {
			remaining := value.Filter[len(str):]
			for i := 0; i < len(remaining); i += 70 {
				epos := i + 70
				if epos > len(remaining) {
					epos = len(remaining)
					fmt.Printf("|%65s %-70s|\n", " ", remaining[i:epos])
				} else {
					str = remaining[i:epos]
					idx := strings.LastIndex(str, " ")
					fmt.Printf("|%65s %-70s|\n", " ", str[:idx])
					i -= (70 - idx)
				}
			}
		}
	}
	fmt.Println("+-------+--------+-------+------+---------------------------------+----------------------------------------------------------------------+")
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

	str = str[:idx] + " " + str[epos:]
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

// count number of '\n'
func countLines(file *os.File) (int, error) {
	reader := bufio.NewReader(file)
	buf := make([]byte, 32*1024)
	lineSep := []byte{'\n'}
	lineCounts := 0
	for {
		c, err := reader.Read(buf)
		lineCounts += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return lineCounts, err

		case err != nil:
			return lineCounts, err
		}
	}
}
