// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/simagix/gox"
)

// ParseLog - parses text message before v4.4
func (li *LogInfo) ParseLog(str string) (LogStats, error) {
	var err error
	var stat LogStats
	matched := regexp.MustCompile(li.regex)

	scan := ""
	aggStages := ""
	if matched.MatchString(str) == true {
		if strings.Index(str, "COLLSCAN") >= 0 {
			scan = COLLSCAN
		}
		if li.collscan == true && scan != COLLSCAN {
			return stat, err
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
		if strings.HasPrefix(ns, "admin.") || strings.HasPrefix(ns, "config.") || strings.HasPrefix(ns, "local.") {
			stat.op = dollarCmd
			return stat, errors.New("system database")
		} else if strings.HasSuffix(ns, ".$cmd") {
			stat.op = dollarCmd
			return stat, errors.New("system command")
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
				return stat, err
			}
			op = res[1]
			filter = res[2]
		}

		if op == "insert" {
			filter = "{ }"
		} else if hasFilter(op) == false {
			return stat, err
		}
		if op == "delete" && strings.Index(filter, "writeConcern:") >= 0 {
			return stat, err
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
			for _, mstr := range []string{"pipeline: [ { $match: ", "pipeline: [ { $sort: ", "$facet: "} {
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
					if mstr == "$facet: " {
						filter = "{ $facet: " + s + " }"
					}
					break
				}
			}
			if s == "" {
				if scan == "COLLSCAN" { // it's a collection scan without $match or $sort
					filter = "{}"
				} else {
					return stat, err
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
					return stat, err
				}
			} else {
				return stat, err
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
		} else if strings.Index(str, "exception: shard version not ok") > 0 {
			return stat, err
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
		filter += aggStages
		milli, _ := strconv.Atoi(ms)
		stat = LogStats{filter: filter, index: index, milli: milli, ns: ns, op: op, scan: scan}
		return stat, nil
	}
	return stat, errors.New("unrecognized log")
}

func getDocByField(str string, key string) string {
	ml := gox.NewMongoLog(str)
	return ml.Get(key)
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
