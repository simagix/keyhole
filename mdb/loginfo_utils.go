// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/simagix/gox"
)

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
			if idx > 0 {
				str = value.Filter[:idx]
			}
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
