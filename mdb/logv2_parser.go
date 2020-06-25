// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"

	"github.com/simagix/gox"
)

var ops = []string{cmdAggregate, cmdDelete, cmdFind, cmdInsert, cmdUpdate}

const cmdAggregate = "aggregate"
const cmdDelete = "delete"
const cmdFind = "find"
const cmdInsert = "insert"
const cmdRemove = "remove"
const cmdUpdate = "update"

// ParseLogv2 - parses text message before v4.4
func (li *LogInfo) ParseLogv2(str string) (LogStats, error) {
	var stat = LogStats{}
	var doc map[string]interface{}
	if strings.Index(str, "durationMillis") < 0 {
		return stat, errors.New("no durationMillis found")
	}
	json.Unmarshal([]byte(str), &doc)
	attr := doc["attr"].(map[string]interface{})
	stat.milli = int(attr["durationMillis"].(float64))
	if attr["ns"] != nil {
		stat.ns = attr["ns"].(string)
	} else if attr["namespace"] != nil {
		stat.ns = attr["namespace"].(string)
	} else {
		return stat, errors.New("no namespace found")
	}
	if strings.HasPrefix(stat.ns, "admin.") || strings.HasPrefix(stat.ns, "config.") || strings.HasPrefix(stat.ns, "local.") {
		stat.op = dollarCmd
		return stat, errors.New("system database")
	} else if strings.HasSuffix(stat.ns, ".$cmd") {
		stat.op = dollarCmd
		return stat, errors.New("system command")
	}
	if attr["planSummary"] != nil {
		plan := attr["planSummary"].(string)
		if plan == COLLSCAN {
			stat.scan = attr["planSummary"].(string)
		} else if strings.HasPrefix(plan, "IXSCAN") {
			stat.index = plan[len("IXSCAN")+1:]
		} else {
			stat.index = plan
		}
	}
	if li.collscan == true && stat.scan != COLLSCAN {
		return stat, nil
	}
	if attr["command"] == nil {
		return stat, errors.New("no command found")
	}
	command := attr["command"].(map[string]interface{})
	if attr["type"] != nil {
		stat.op = attr["type"].(string)
	}
	if stat.op == "command" {
		stat.op = ""
		for _, v := range ops {
			if command[v] != nil {
				stat.op = v
				break
			}
		}
		if stat.op == cmdFind {
			if command["filter"] == nil {
				stat.filter = "{}"
			} else {
				fmap := command["filter"].(map[string]interface{})
				if isRegex(fmap) == false {
					walker := gox.NewMapWalker(cb)
					doc := walker.Walk(fmap)
					if buf, err := json.Marshal(doc); err == nil {
						stat.filter = string(buf)
					} else {
						stat.filter = "{}"
					}
				} else {
					buf, _ := json.Marshal(fmap)
					str := string(buf)
					re := regexp.MustCompile(`{(.*):{"\$regularExpression":{"options":"(\S+)?","pattern":"(\^)?(\S+)"}}}`)
					stat.filter = re.ReplaceAllString(str, "{$1:/$3.../$2}")
				}
			}
		} else if stat.op == "" {
			return stat, errors.New("no op found")
		}
	}
	if stat.op == cmdInsert {
		stat.filter = "N/A"
	} else if (stat.op == cmdUpdate || stat.op == cmdRemove || stat.op == cmdDelete) && stat.filter == "" {
		walker := gox.NewMapWalker(cb)
		doc := walker.Walk(command["q"].(map[string]interface{}))
		if buf, err := json.Marshal(doc); err == nil {
			stat.filter = string(buf)
		} else {
			stat.filter = "{}"
		}
	} else if stat.op == cmdAggregate {
		pipeline := command["pipeline"].([]interface{})
		var stage interface{}
		for _, v := range pipeline {
			stage = v
			break
		}
		fmap := stage.(map[string]interface{})
		if isRegex(fmap) == false {
			walker := gox.NewMapWalker(cb)
			doc := walker.Walk(fmap)
			if buf, err := json.Marshal(doc); err == nil {
				stat.filter = string(buf)
			} else {
				stat.filter = "{}"
			}
			if strings.Index(stat.filter, "$match") < 0 && strings.Index(stat.filter, "$sort") < 0 &&
				strings.Index(stat.filter, "$facet") < 0 {
				stat.filter = "{}"
			}
		} else {
			buf, _ := json.Marshal(fmap)
			str := string(buf)
			re := regexp.MustCompile(`{(.*):{"\$regularExpression":{"options":"(\S+)?","pattern":"(\^)?(\S+)"}}}`)
			stat.filter = re.ReplaceAllString(str, "{$1:/$3.../$2}")
		}
	}
	if stat.op == "" {
		return stat, nil
	}
	re := regexp.MustCompile(`\[(1,)*1\]`)
	stat.filter = re.ReplaceAllString(stat.filter, `[...]`)
	re = regexp.MustCompile(`^{("\$match"|"\$sort"):(\S+)}$`)
	stat.filter = re.ReplaceAllString(stat.filter, `$2`)
	re = regexp.MustCompile(`^{("(\$facet")):\S+}$`)
	stat.filter = re.ReplaceAllString(stat.filter, `{$1:...}`)
	re = regexp.MustCompile(`{"\$oid":1}`)
	stat.filter = re.ReplaceAllString(stat.filter, `1`)
	return stat, nil
}

func isRegex(doc map[string]interface{}) bool {
	if buf, err := json.Marshal(doc); err != nil {
		return false
	} else if strings.Index(string(buf), "$regularExpression") >= 0 {
		return true
	}
	return false
}

func cb(value interface{}) interface{} {
	return 1
}
