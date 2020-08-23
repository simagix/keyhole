// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/simagix/gox"
)

// Logv2 stores logv2 info
type Logv2 struct {
	Attributes struct {
		Command            map[string]interface{} `json:"command" bson:"command"`
		Milli              int                    `json:"durationMillis" bson:"durationMillis"`
		NS                 string                 `json:"ns" bson:"ns"`
		OriginatingCommand map[string]interface{} `json:"originatingCommand" bson:"originatingCommand"`
		PlanSummary        string                 `json:"planSummary" bson:"planSummary"`
		Type               string                 `json:"type" bson:"type"`
	} `json:"attr" bson:"attr"`
	Component string    `json:"c" bson:"c"`
	ID        int       `json:"id" bson:"id"`
	Message   string    `json:"msg" bson:"msg"`
	Severity  string    `json:"s" bson:"s"`
	Timestamp time.Time `json:"t" bson:"t"`
}

var ops = []string{cmdAggregate, cmdDelete, cmdFind, cmdGetMore, cmdInsert, cmdUpdate}

const cmdAggregate = "aggregate"
const cmdCreateIndexes = "createIndexes"
const cmdDelete = "delete"
const cmdFind = "find"
const cmdGetMore = "getMore"
const cmdInsert = "insert"
const cmdRemove = "remove"
const cmdUpdate = "update"

// ParseLogv2 - parses text message before v4.4
func (li *LogInfo) ParseLogv2(str string) (LogStats, error) {
	var err error
	var stat = LogStats{}
	var doc Logv2
	if strings.LastIndex(str, "durationMillis") < 0 {
		return stat, errors.New("no durationMillis found")
	}
	// if err = json.Unmarshal([]byte(str), &doc); err != nil {
	// 	return stat, err
	// }
	bson.UnmarshalExtJSON([]byte(str), false, &doc)
	c := doc.Component
	if c != "COMMAND" && c != "QUERY" && c != "WRITE" {
		return stat, errors.New("unsupported command")
	}
	stat.milli = doc.Attributes.Milli
	if doc.Attributes.NS == "" {
		return stat, errors.New("no namespace found")
	}
	stat.ns = doc.Attributes.NS
	if stat.ns == "" {
		return stat, errors.New("no ns info")
	} else if strings.HasPrefix(stat.ns, "admin.") || strings.HasPrefix(stat.ns, "config.") || strings.HasPrefix(stat.ns, "local.") {
		stat.op = dollarCmd
		return stat, errors.New("system database")
	} else if strings.HasSuffix(stat.ns, ".$cmd") {
		stat.op = dollarCmd
		return stat, errors.New("system command")
	}

	if doc.Attributes.PlanSummary != "" { // not insert
		plan := doc.Attributes.PlanSummary
		if plan == COLLSCAN {
			stat.scan = COLLSCAN
		} else if strings.HasPrefix(plan, "IXSCAN") {
			stat.index = plan[len("IXSCAN")+1:]
		} else {
			stat.index = plan
		}
	}

	if li.Collscan == true && stat.scan != COLLSCAN {
		return stat, errors.New("skip, -collscan")
	}
	if doc.Attributes.Command == nil {
		return stat, errors.New("no command found")
	}
	command := doc.Attributes.Command
	stat.op = doc.Attributes.Type
	if stat.op == "command" || stat.op == "none" {
		stat.op = getOp(command)
	}
	var isGetMore bool
	if stat.op == cmdGetMore {
		isGetMore = true
		command = doc.Attributes.OriginatingCommand
		stat.op = getOp(command)
	}
	if stat.op == cmdFind {
		if command["filter"] == nil {
			return stat, errors.New("no filter found")
		}
		fmap := command["filter"].(map[string]interface{})
		if isRegex(fmap) == false {
			walker := gox.NewMapWalker(cb)
			doc := walker.Walk(fmap)
			var data []byte
			if data, err = json.Marshal(doc); err != nil {
				return stat, err
			}
			stat.filter = string(data)
			if stat.filter == `{"":null}` {
				stat.filter = "{}"
			}
		} else {
			buf, _ := json.Marshal(fmap)
			str := string(buf)
			re := regexp.MustCompile(`{(.*):{"\$regularExpression":{"options":"(\S+)?","pattern":"(\^)?(\S+)"}}}`)
			stat.filter = re.ReplaceAllString(str, "{$1:/$3.../$2}")
		}
	} else if stat.op == cmdInsert || stat.op == cmdCreateIndexes {
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
		pipeline := command["pipeline"].(primitive.A)
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
				strings.Index(stat.filter, "$facet") < 0 && strings.Index(stat.filter, "$indexStats") < 0 {
				stat.filter = "{}"
			}
		} else {
			buf, _ := json.Marshal(fmap)
			str := string(buf)
			re := regexp.MustCompile(`{(.*):{"\$regularExpression":{"options":"(\S+)?","pattern":"(\^)?(\S+)"}}}`)
			stat.filter = re.ReplaceAllString(str, "{$1:/$3.../$2}")
		}
	} else if li.verbose == true {
		fmt.Println(stat.op, str)
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
	if isGetMore {
		stat.op = cmdGetMore
	}
	utc := doc.Timestamp.Format(time.RFC3339)[:15] + `0:00Z`
	stat.utc = utc
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

func getOp(command map[string]interface{}) string {
	for _, v := range ops {
		if command[v] != nil {
			return v
		}
	}
	return ""
}

func cb(value interface{}) interface{} {
	return 1
}
