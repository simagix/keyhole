package stats

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
)

// OpPattern -
type OpPattern struct {
	Command   string
	Namespace string
	Filter    string
	Milli     int
	Count     int
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
		if err != nil {
			break
		} else if matched.MatchString(string(buf)) == true {
			s := string(buf)
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

			re = regexp.MustCompile(`(createIndexes: "\w+", |find: "\w+", |, \$db: "\w+" |,? ?skip: \d+|, limit: \d+|, batchSize: \d+|, singleBatch: \w+)|, multi: \w+|, upsert: \w+|, ordered: \w+`)
			filter = re.ReplaceAllString(filter, "")
			re = regexp.MustCompile(`(: "[^"]*"|: \d+)`)
			filter = re.ReplaceAllString(filter, ": 1")
			key := op + "." + filter
			_, ok := opMap[key]
			m, _ := strconv.Atoi(ms)
			if ok {
				x := opMap[key].Milli + m
				y := opMap[key].Count + 1
				opMap[key] = OpPattern{Command: opMap[key].Command, Namespace: ns, Filter: opMap[key].Filter, Milli: x, Count: y}
			} else {
				opMap[key] = OpPattern{Command: op, Namespace: ns, Filter: filter, Milli: m, Count: 1}
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
	fmt.Println("+-------------+---------+------+------------------------------+----------------------------------------------------------------------+")
	fmt.Printf("| Command     | Time ms | Count| %-29s| %-69s|\n", "Namespace", "Query Pattern")
	fmt.Println("|-------------+---------+------+------------------------------+----------------------------------------------------------------------|")
	for _, value := range arr {
		fmt.Printf("|%-13s|%9.1f|%6d|%-30s|%-70s|\n", value.Command, float64(value.Milli)/float64(value.Count), value.Count, value.Namespace, value.Filter)
	}
	fmt.Println("+-------------+---------+------+------------------------------+----------------------------------------------------------------------+")
}
