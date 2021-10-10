// Copyright 2020 - present, Kuei-chun Chen. All rights reserved.

package mdb

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/simagix/gox"
	"go.mongodb.org/mongo-driver/bson"
)

// LogVersion2 stores logv2 info
type LogVersion2 struct {
	Attributes struct {
		Command            bson.M `json:"command" bson:"command"`
		Milli              int    `json:"durationMillis" bson:"durationMillis"`
		NS                 string `json:"ns" bson:"ns"`
		OriginatingCommand bson.M `json:"originatingCommand" bson:"originatingCommand"`
		PlanSummary        string `json:"planSummary" bson:"planSummary"`
		Type               string `json:"type" bson:"type"`
	} `json:"attr" bson:"attr"`
	Component string             `json:"c" bson:"c"`
	Context   string             `json:"ctx" bson:"ctx"`
	ID        int                `json:"id" bson:"id"`
	Message   string             `json:"msg" bson:"msg"`
	Severity  string             `json:"s" bson:"s"`
	DateTime  primitive.DateTime `json:"t" bson:"t"`
}

// OutputLogInOldFormat view v4.4+ logs in old format
func OutputLogInOldFormat(filename string) error {
	var err error
	var doc LogVersion2
	var file *os.File
	var reader *bufio.Reader
	var buf []byte

	if file, err = os.Open(filename); err != nil {
		return err
	}
	defer file.Close()
	if reader, err = gox.NewReader(file); err != nil {
		return err
	}
	for {
		if buf, _, err = reader.ReadLine(); err != nil { // 0x0A separator = newline
			break
		}
		if len(buf) == 0 {
			continue
		}
		if err = bson.UnmarshalExtJSON(buf, false, &doc); err != nil {
			fmt.Println(err)
			continue
		}
		attr := doc.Attributes
		message := ""
		if attr.NS != "" {
			message = fmt.Sprintf(`%v %v PlanSummary: %v %v %v %vms`, attr.Type, attr.NS, attr.PlanSummary,
				gox.Stringify(attr.Command), gox.Stringify(attr.OriginatingCommand), attr.Milli)
		}
		result := fmt.Sprintf(`%v %v %v [%v] %v %v`, doc.DateTime.Time().Format(time.RFC3339),
			doc.Severity, doc.Component, doc.Context, doc.Message, message)
		fmt.Println(result)
	}
	return err
}
