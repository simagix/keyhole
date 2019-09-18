// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/simagix/gox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// QueryExplainerWrapper calls queryPlanner and cardinality
func QueryExplainerWrapper(client *mongo.Client, filename string, verbose bool) error {
	var err error
	qe := NewQueryExplainer(client)
	qe.SetVerbose(verbose)
	if err = qe.ReadQueryShapeFromFile(filename); err != nil {
		return err
	}
	card := NewCardinality(client)
	card.SetVerbose(verbose)
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
	document := make(map[string]interface{})
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
	stdout := strings.Join(strs, "\n")
	document["stdout"] = stdout
	fmt.Println(stdout)
	ofile := filepath.Base(filename) + "-explain.json.gz"
	if err = gox.OutputGzipped([]byte(gox.Stringify(document)), ofile); err != nil {
		return err
	}
	fmt.Println("Explain JSON and output written to", ofile)
	return err
}

// PrintExplainResults prints explain results
func PrintExplainResults(filename string) error {
	var err error
	var data []byte
	if strings.HasSuffix(filename, ".json.gz") {
		var file *os.File
		var gz *gzip.Reader
		if file, err = os.Open(filename); err != nil {
			return err
		}
		if gz, err = gzip.NewReader(file); err != nil {
			return err
		}
		if data, err = ioutil.ReadAll(gz); err != nil {
			return err
		}
	} else if data, err = ioutil.ReadFile(filename); err != nil {
		return err
	}
	doc := bson.M{}
	json.Unmarshal(data, &doc)
	if doc["stdout"] == nil {
		usage := "Usage: keyhole --explain <mongod.log> <uri> | <result.json.gz>"
		return errors.New(usage)
	}
	fmt.Println(doc["stdout"])
	return err
}
