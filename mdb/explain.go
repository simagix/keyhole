// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/simagix/gox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Explain stores explain object info
type Explain struct {
	verbose bool
}

// NewExplain returns Explain struct
func NewExplain() *Explain {
	return &Explain{}
}

// SetVerbose sets verbosity
func (e *Explain) SetVerbose(verbose bool) {
	e.verbose = verbose
}

// ExecuteAllPlans calls queryPlanner and cardinality
func (e *Explain) ExecuteAllPlans(client *mongo.Client, filename string) error {
	var err error
	var reader *bufio.Reader

	if reader, err = gox.NewFileReader(filename); err != nil {
		return err
	}
	qe := NewQueryExplainer(client)
	qe.SetVerbose(e.verbose)
	card := NewCardinality(client)
	card.SetVerbose(e.verbose)
	stdout := ""
	counter := 0
	for {
		buffer, _, rerr := reader.ReadLine()
		if rerr != nil {
			break
		} else if !strings.HasSuffix(string(buffer), "ms") {
			continue
		}
		if err = qe.ReadQueryShape(buffer); err != nil {
			continue
		}
		var summary CardinalitySummary
		keys := GetKeys(qe.ExplainCmd.Filter)
		keys = append(keys, GetKeys(qe.ExplainCmd.Sort)...)
		pos := strings.Index(qe.NameSpace, ".")
		db := qe.NameSpace[:pos]
		collection := qe.NameSpace[pos+1:]
		if summary, err = card.GetCardinalityArray(db, collection, keys); err != nil {
			return err
		}
		var explainSummary ExplainSummary
		if explainSummary, err = qe.Explain(); err != nil {
			fmt.Println(err.Error())
		}
		strs := []string{}
		strs = append(strs, qe.GetSummary(explainSummary))
		strs = append(strs, "=> All Applicable Indexes Scores")
		strs = append(strs, "=========================================")
		scores := qe.GetIndexesScores(keys)
		strs = append(strs, gox.Stringify(scores, "", "  "))
		strs = append(strs, card.GetSummary(summary)+"\n")
		document := bson.M{}
		document["ns"] = qe.NameSpace
		document["cardinality"] = summary
		document["explain"] = explainSummary
		document["scores"] = scores
		if len(summary.List) > 0 {
			recommendedIndex := GetIndexSuggestion(qe.ExplainCmd, summary.List)
			document["recommendedIndex"] = recommendedIndex
			strs = append(strs, "Index Suggestion:", gox.Stringify(recommendedIndex))
		}
		strs = append(strs, "")
		stdout = strings.Join(strs, "\n")
		document["stdout"] = stdout
		counter++
		if counter == 1 {
			fmt.Println(stdout)
		}
		os.Mkdir(outdir, 0755)
		ofile := fmt.Sprintf("%v/%v-explain-%03d.json.gz", outdir, filepath.Base(filename), counter)
		data, _ := bson.MarshalExtJSON(document, false, false)
		if err = gox.OutputGzipped(data, ofile); err != nil {
			return err
		}
		fmt.Println("json data written to", ofile)
	}
	return err
}

// PrintExplainResults prints explain results
func (e *Explain) PrintExplainResults(filename string) error {
	var err error
	var data []byte
	var reader *bufio.Reader

	if reader, err = gox.NewFileReader(filename); err != nil {
		return err
	}
	if data, err = io.ReadAll(reader); err != nil {
		return err
	}
	doc := bson.M{}
	json.Unmarshal(data, &doc)
	if doc["stdout"] == nil {
		usage := "usage: keyhole --explain <mongod.log> <uri> | <result.json.gz>"
		return errors.New(usage)
	}
	fmt.Println(doc["stdout"])
	return err
}
