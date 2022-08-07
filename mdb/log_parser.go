// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/simagix/gox"
)

var hasFilters = map[string]bool{"count": true, "delete": true, "find": true, "remove": true, "update": true, "aggregate": true, "getMore": true, "getmore": true, "findAndModify": true, "distinct": true}

// ParseLog - parses text message before v4.4
func (li *LogInfo) ParseLog(str string) (LogStats, error) {
	var err error
	var stat LogStats

	scan := ""
	aggStages := ""
	if !strings.HasSuffix(str, "ms") {
		return stat, errors.New("unrecognized log")
	}
	if !li.regexp.MatchString(str) {
		return stat, errors.New("unrecognized log")
	}
	if strings.Contains(str, "COLLSCAN") {
		scan = COLLSCAN
	}
	if li.Collscan && scan != COLLSCAN {
		return stat, err
	}
	result := li.regexp.FindStringSubmatch(str)
	c := result[2]
	if c != "COMMAND" && c != "QUERY" && c != "WRITE" {
		return stat, errors.New("unsupported command")
	}
	isFound := false
	bpos := 0 // begin position
	epos := 0 // end position
	body := result[6]
	for _, r := range body {
		epos++
		if !isFound && r == '{' {
			isFound = true
			bpos++
		} else if isFound {
			if r == '{' {
				bpos++
			} else if r == '}' {
				bpos--
			}
		}

		if isFound && bpos == 0 {
			break
		}
	}

	re := regexp.MustCompile(`^(\w+) ({.*})$`)
	utc := result[1][:16] + `:00Z`
	op := result[4]
	ns := result[5]
	if strings.HasPrefix(ns, "admin.") || strings.HasPrefix(ns, "config.") || strings.HasPrefix(ns, "local.") {
		stat.op = dollarCmd
		return stat, errors.New("system database")
		// } else if strings.HasSuffix(ns, ".$cmd") {
		// stat.op = dollarCmd
		// return stat, errors.New("system command")
	}
	filter := body[:epos]
	ms := result[7]
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
	} else if op == "query" {
		op = "find"
	}

	if op == "insert" {
		filter = "{ }"
	} else if !hasFilters[op] {
		return stat, err
	}
	if op == "delete" && strings.Contains(filter, "writeConcern:") {
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
		nstr = strings.TrimSuffix(nstr, " }")
		filter = "{" + nstr + ": 1}"
	} else if op == "delete" || op == "update" || op == "remove" || op == "findAndModify" {
		var s string
		if strings.Contains(filter, "query: ") {
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
			s = getDocByField(filter, mstr)
			if s != "" {
				filter = s
				x := strings.Index(body, "$group: ")
				y := strings.Index(body, "$sort: ")
				if x > 0 && (x < y || y < 0) {
					aggStages = ", group: " + strings.ReplaceAll(getDocByField(ns, "$group: "), "1.0", "1")
				}
				srt := getDocByField(body, "$sort: ")
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
		s := getDocByField(body, "originatingCommand: ")
		if s != "" {
			s = getDocByField(s, "filter: ")
			for _, mstr := range []string{"filter: ", "pipeline: [ { $match: ", "pipeline: [ { $sort: "} {
				s = getDocByField(body, mstr)
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
	} else if strings.Contains(str, "planSummary: EOF") {
		index = "EOF"
	} else if strings.Contains(str, "planSummary: IDHACK") {
		index = "IDHACK"
	} else if strings.Contains(str, "planSummary: COUNT_SCAN") {
		index = "COUNT_SCAN"
	} else if strings.Contains(str, "planSummary: DISTINCT_SCAN") {
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
	re = regexp.MustCompile(`(: \/(\^)?.*?(\$)?\/([igm])?(,)?\s)`)
	filter = re.ReplaceAllString(filter, ": /$2regex$3/$4$5 ")
	filter = strings.Replace(strings.Replace(filter, "{ ", "{", -1), " }", "}", -1)
	filter += aggStages
	milli, _ := strconv.Atoi(ms)

	if strings.HasSuffix(ns, ".$cmd") {
		ns = strings.TrimSuffix(ns, "$cmd")
		coll := getDocByField(body, op+":")
		coll = strings.TrimPrefix(coll, `"`)
		quote := strings.Index(coll, `"`)
		ns += coll[:quote]
	}
	reslen := getDocByField(str, "reslen:")
	resLength := 0
	idx := strings.Index(reslen, " ")
	if reslen != "" && idx > 0 {
		resLength = ToInt(reslen[:idx])
	}
	stat = LogStats{filter: filter, index: index, milli: milli, ns: ns, op: op,
		reslen: resLength, scan: scan, utc: utc}
	return stat, nil
}

func getDocByField(str string, key string) string {
	ml := gox.NewMongoLog(str)
	return ml.Get(key)
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
