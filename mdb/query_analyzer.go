// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// QueryAnalyzer stores query analyzer info
type QueryAnalyzer struct {
	client   *mongo.Client
	database string
	filter   bson.M
	verbose  bool
}

// NewQueryAnalyzer returns QueryAnalyzer
func NewQueryAnalyzer(client *mongo.Client) *QueryAnalyzer {
	return &QueryAnalyzer{client: client}
}

// SetDatabase sets database
func (qa *QueryAnalyzer) SetDatabase(database string) {
	qa.database = database
}

// GetFilter returns filter
func (qa *QueryAnalyzer) GetFilter() bson.M {
	return qa.filter
}

// SetFilter returns filter
func (qa *QueryAnalyzer) SetFilter(filter bson.M) {
	qa.filter = filter
}

// SetVerbose sets verbosity
func (qa *QueryAnalyzer) SetVerbose(verbose bool) {
	qa.verbose = verbose
}

// Explain explains query plans
func (qa *QueryAnalyzer) Explain(collectionName string, filter map[string]interface{}) (bson.M, error) {
	qa.filter = filter
	var err error
	ctx := context.Background()
	command := bson.M{"explain": bson.M{"count": collectionName, "query": filter}}
	var result = bson.M{}
	for i := 0; i < 3; i++ { // a hack to avoid error (CommandNotFound) Explain failed due to unknown command: query
		err = qa.client.Database(qa.database).RunCommand(ctx, command).Decode(&result)
		if err == nil {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	return result, err
}

// GetSummary get summary of explain executionStats
func (qa *QueryAnalyzer) GetSummary(doc bson.M) (string, error) {
	var err error
	var buffer bytes.Buffer
	var shardName string
	buffer.WriteString("Explain:\n")
	buffer.WriteString("--------------------------------------------------------------------------------\n")

	// find queryPlanner.winningPlan
	winningPlan := doc["queryPlanner"].(bson.M)["winningPlan"].(bson.M)
	if winningPlan["shards"] != nil {
		shardName = winningPlan["shards"].(primitive.A)[0].(bson.M)["shardName"].(string)
		winningPlan = winningPlan["shards"].(primitive.A)[0].(bson.M)["winningPlan"].(bson.M)
	}

	inputStage := winningPlan["inputStage"].(bson.M)
	if shardName == "" {
		buffer.WriteString("Replica Set\n")
	} else {
		buffer.WriteString("Shard Name: " + shardName + "\n")
		inputStage = inputStage["inputStage"].(bson.M)
	}

	buffer.WriteString(fmt.Sprintf("* filter: %v\n", Stringify(qa.filter)))
	buffer.WriteString("* Winning Plan\n")
	if inputStage["keyPattern"] != nil {
		buffer.WriteString(fmt.Sprintf("  - key pattern: %v\n", Stringify(inputStage["keyPattern"].(bson.M))))
	}
	buffer.WriteString(fmt.Sprintf("  - stage: %v\n", inputStage["stage"]))

	// find executionStats.allPlansExecution
	allPlansExecution := doc["executionStats"].(bson.M)["allPlansExecution"].(primitive.A)
	if len(allPlansExecution) > 0 && allPlansExecution[0].(bson.M)["allPlans"] != nil {
		allPlansExecution = allPlansExecution[0].(bson.M)["allPlans"].(primitive.A)
	}

	if len(allPlansExecution) == 0 {
		return buffer.String(), err
	}

	buffer.WriteString("--------------------------------------------------------------------------------\n")
	buffer.WriteString("* All Plans Execution\n")
	for _, execution := range allPlansExecution {
		executionStages := execution.(bson.M)["executionStages"].(bson.M)
		buffer.WriteString(fmt.Sprintf(" ├─ totalKeysExamined: %v\n", execution.(bson.M)["totalKeysExamined"]))
		buffer.WriteString(fmt.Sprintf(" └─ totalDocsExamined: %v\n", execution.(bson.M)["totalDocsExamined"]))
		buffer.WriteString(fmt.Sprintf("   ├─ executionStages.stage: %v\n", executionStages["stage"]))
		buffer.WriteString(fmt.Sprintf("   ├─ executionStages.executionTimeMillisEstimate: %v\n", executionStages["executionTimeMillisEstimate"]))
		buffer.WriteString(fmt.Sprintf("   ├─ executionStages.needYield: %v\n", executionStages["needYield"]))
		buffer.WriteString(fmt.Sprintf("   ├─ executionStages.restoreState: %v\n", executionStages["restoreState"]))
		buffer.WriteString(fmt.Sprintf("   └─ executionStages.saveState: %v\n", executionStages["saveState"]))

		var inputStages primitive.A
		if executionStages["inputStage"] != nil {
			inputStages = append(inputStages, executionStages["inputStage"].(bson.M))
		} else if executionStages["inputStages"] != nil {
			inputStages = executionStages["inputStages"].(primitive.A)
		}

		for _, input := range inputStages {
			if input.(bson.M)["inputStage"] != nil {
				input = input.(bson.M)["inputStage"].(bson.M)
			}
			inputStage = input.(bson.M)
			buffer.WriteString(fmt.Sprintf("     ├─ key pattern: %v\n", Stringify(inputStage["keyPattern"].(bson.M))))
			buffer.WriteString(fmt.Sprintf("     ├─ inputStage.stage: %v\n", inputStage["stage"]))
			buffer.WriteString(fmt.Sprintf("     ├─ inputStage.executionTimeMillisEstimate: %v\n", inputStage["executionTimeMillisEstimate"]))
			buffer.WriteString(fmt.Sprintf("     ├─ inputStage.needYield: %v\n", inputStage["needYield"]))
			buffer.WriteString(fmt.Sprintf("     ├─ inputStage.restoreState: %v\n", inputStage["restoreState"]))
			buffer.WriteString(fmt.Sprintf("     ├─ inputStage.saveState: %v\n", inputStage["saveState"]))
			buffer.WriteString(fmt.Sprintf("     └─ inputStage.works: %v\n", inputStage["works"]))
		}

		buffer.WriteString("\n")
	}

	buffer.WriteString("--------------------------------------------------------------------------------\n")
	return buffer.String(), err
}

// OutputGzipped writes doc to a gzipped file
func (qa *QueryAnalyzer) OutputGzipped(doc bson.M, filename string) error {
	var zbuf bytes.Buffer
	gz := gzip.NewWriter(&zbuf)
	gz.Write([]byte(Stringify(doc)))
	gz.Close() // close this before flushing the bytes to the buffer.
	return ioutil.WriteFile(filename, zbuf.Bytes(), 0644)
}
