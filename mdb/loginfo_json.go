// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
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

// parseJSON - parses text message before v4.4
func (li *LogInfo) parseJSON(str string) (logStats, error) {
	var stat = logStats{}
	var doc map[string]interface{}
	if strings.Index(str, "durationMillis") < 0 {
		return stat, errors.New("no durationMillis found")
	}
	json.Unmarshal([]byte(str), &doc)
	attr := doc["attr"].(map[string]interface{})
	stat.milli = int(attr["durationMillis"].(float64))
	stat.ns = attr["ns"].(string)
	if strings.HasPrefix(stat.ns, "admin.") || strings.HasPrefix(stat.ns, "config.") || strings.HasSuffix(stat.ns, ".$cmd") {
		return stat, errors.New("system database")
	}
	if attr["planSummary"] != nil {
		plan := attr["planSummary"].(string)
		if plan == COLLSCAN {
			stat.scan = attr["planSummary"].(string)
		} else if strings.HasPrefix(plan, "IXSCAN") {
			stat.index = plan[len("IXSCAN")+1:]
		}
	}
	command := attr["command"].(map[string]interface{})
	if attr["type"] != nil {
		stat.op = attr["type"].(string)
		if (stat.op == cmdRemove || stat.op == cmdUpdate) && stat.scan != "" {
			walker := gox.NewMapWalker(cb)
			doc := walker.Walk(command["q"].(map[string]interface{}))
			if buf, err := json.Marshal(doc); err == nil {
				stat.filter = string(buf)
			} else {
				stat.filter = "{}"
			}
		}
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
		} else if stat.op == "" {
			return stat, errors.New("no op found")
		}
	}
	if stat.op == cmdInsert {
		stat.filter = "N/A"
	} else if (stat.op == cmdUpdate || stat.op == cmdRemove || stat.op == cmdDelete) && stat.filter == "" {
		stat.filter = "{}"
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
		fmt.Println(gox.Stringify(doc, "", "  "))
		os.Exit(0)
	}
	re := regexp.MustCompile(`:\[(\S+)\]`)
	stat.filter = re.ReplaceAllString(stat.filter, ":[...]")
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
