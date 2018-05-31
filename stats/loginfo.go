package stats

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// OpPattern -
type OpPattern struct {
	Command   string
	Namespace string
	Filter    string
	Milli     int
	Count     int
	Scan      string
}

// LogInfo -
func LogInfo(filename string) {
	var opMap map[string]OpPattern
	opMap = make(map[string]OpPattern)
	var matched = regexp.MustCompile(`^\S+ .? (\w+)\s+\[\w+\] (\w+) (\S+) \S+: (.*) (\d+)ms$`)
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println("error opening file ", err)
		return
	}
	defer f.Close()
	r := bufio.NewReader(f)
	for {
		buf, _, err := r.ReadLine() // 0x0A separator = newline
		scan := ""
		if err != nil {
			break
		} else if matched.MatchString(string(buf)) == true {
			s := string(buf)
			if strings.Index(s, "COLLSCAN") >= 0 {
				scan = "COLLSCAN"
			}
			result := matched.FindStringSubmatch(s)
			b := false
			n := 0
			e := 0
			for _, r := range result[4] {
				e++
				if b == false && r == '{' {
					b = true
					n++
				} else if b == true {
					if r == '{' {
						n++
					} else if r == '}' {
						n--
					}
				}

				if b == true && n == 0 {
					break
				}
			}

			re := regexp.MustCompile(`^(\w+) ({.*})$`)
			op := result[2]
			ns := result[3]
			filter := result[4][:e]
			ms := result[5]
			if op == "command" {
				res := re.FindStringSubmatch(filter)
				if len(res) < 3 {
					continue
				}
				op = res[1]
				filter = res[2]
			}

			if op == "insert" || op == "moveChunk" || op == "splitVector" {
				continue
			}

			re = regexp.MustCompile(`(createIndexes: "\w+", |find: "\w+", |, \$db: "\w+" |,? ?skip: \d+|, limit: \d+|, batchSize: \d+|, singleBatch: \w+)|, multi: \w+|, upsert: \w+|, ordered: \w+|, shardVersion: \[ Timestamp 0\|0, ObjectId\('\S+'\) \]`)
			filter = re.ReplaceAllString(filter, "")
			re = regexp.MustCompile(`(: "[^"]*"|: \d+|: new Date\(\d+?\))`)
			filter = re.ReplaceAllString(filter, ": 1")
			re = regexp.MustCompile(`(ObjectId\('\S+'\))`)
			filter = re.ReplaceAllString(filter, "ObjectId(1)")
			re = regexp.MustCompile(`(q: {.*}), u: {.*}`)
			filter = re.ReplaceAllString(filter, "$1")
			filter = removeInElements(filter)
			key := op + "." + filter
			_, ok := opMap[key]
			m, _ := strconv.Atoi(ms)
			if ok {
				x := opMap[key].Milli + m
				y := opMap[key].Count + 1
				opMap[key] = OpPattern{Command: opMap[key].Command, Namespace: ns, Filter: opMap[key].Filter, Milli: x, Count: y, Scan: scan}
			} else {
				opMap[key] = OpPattern{Command: op, Namespace: ns, Filter: filter, Milli: m, Count: 1, Scan: scan}
			}
		}
	}

	arr := make([]OpPattern, 0, len(opMap))
	for _, value := range opMap {
		arr = append(arr, value)
	}
	sort.Slice(arr, func(i, j int) bool {
		return float64(arr[i].Milli)/float64(arr[i].Count) > float64(arr[j].Milli)/float64(arr[j].Count)
	})
	fmt.Println("+-------------+--------+-------+------+---------------------------------+----------------------------------------------------------------------+")
	fmt.Printf("| Command     |COLLSCAN| avg ms| Count| %-32s| %-69s|\n", "Namespace", "Query Pattern")
	fmt.Println("|-------------+--------+-------+------+---------------------------------+----------------------------------------------------------------------|")
	for _, value := range arr {
		str := value.Filter
		if len(str) > 70 {
			str = value.Filter[:70]
		}
		if len(value.Command) > 13 {
			value.Command = value.Command[:13]
		}
		if len(value.Namespace) > 33 {
			value.Namespace = value.Namespace[:30] + "..."
		}
		fmt.Printf("|%-13s|%8s|%7.0f|%6d|%-33s|%-70s|\n", value.Command, value.Scan, float64(value.Milli)/float64(value.Count), value.Count, value.Namespace, str)
		if len(value.Filter) > 70 {
			remaining := value.Filter[70:]
			for i := 0; i < len(remaining); i += 70 {
				e := i + 70
				if e > len(remaining) {
					e = len(remaining)
				}
				fmt.Printf("|%71s|%-70s|\n", " ", remaining[i:e])
			}
		}
	}
	fmt.Println("+-------------+--------+-------+------+---------------------------------+----------------------------------------------------------------------+")
}

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
