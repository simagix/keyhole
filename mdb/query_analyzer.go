// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/simagix/gox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// QueryAnalyzer stores query analyzer info
type QueryAnalyzer struct {
	client     *mongo.Client
	database   string
	collection string
	filter     map[string]interface{}
	sort       map[string]interface{}
	verbose    bool
}

type inputStagesLevel struct {
	inputStages primitive.A
	level       int
}

// StageStats stores stats for each stage
type StageStats struct {
	Level             int
	Score             float64         `json:"Score"`
	Stage             string          `json:"stage"`
	Filter            *gox.OrderedMap `json:"filter"`
	KeyPattern        *gox.OrderedMap `json:"keyPattern"`
	Advanced          int32           `json:"advanced"`
	Works             int32           `json:"works"`
	ExecTimeMillisEst int32           `json:"executionTimeMillisEstimate"`
	TotalKeysExamined int32           `json:"totalKeysExamined"`
	TotalDocsExamined int32           `json:"totalDocsExamined"`
	InputStages       []StageStats    `json:"inputStages"`
}

// ExplainSummary stores explain summary
type ExplainSummary struct {
	ShardName              string       `json:"shardName"`
	ExecutionStats         StageStats   `json:"executionStats"`
	AllPlansExecutionStats []StageStats `json:"allPlansExecution"`
}

// NewQueryAnalyzer returns QueryAnalyzer
func NewQueryAnalyzer(client *mongo.Client) *QueryAnalyzer {
	return &QueryAnalyzer{client: client}
}

// SetDatabase sets database
func (qa *QueryAnalyzer) SetDatabase(database string) {
	qa.database = database
}

// SetCollection sets collection name
func (qa *QueryAnalyzer) SetCollection(collection string) {
	qa.collection = collection
}

// SetVerbose sets verbosity
func (qa *QueryAnalyzer) SetVerbose(verbose bool) {
	qa.verbose = verbose
}

// GetFilter sets verbosity
func (qa *QueryAnalyzer) GetFilter() map[string]interface{} {
	return qa.filter
}

// Explain explains query plans
func (qa *QueryAnalyzer) Explain() (ExplainSummary, error) {
	var err error
	ctx := context.Background()
	fmt.Println(qa)
	command := bson.M{"explain": bson.M{"find": qa.collection, "filter": qa.filter}}
	if qa.sort != nil {
		command["explain"].(bson.M)["sort"] = qa.sort
	}
	var doc = bson.M{}
	for i := 0; i < 3; i++ { // a hack to avoid error (CommandNotFound) Explain failed due to unknown command: query
		err = qa.client.Database(qa.database).RunCommand(ctx, command).Decode(&doc)
		if err == nil {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if err != nil {
		return ExplainSummary{}, err
	}
	winStage := doc["queryPlanner"].(bson.M)["winningPlan"].(bson.M)["stage"].(string)
	if winStage == "EOF" {
		return ExplainSummary{}, errors.New("no data found to be explained")
	} else if winStage == "COLLSCAN" {
		return ExplainSummary{}, errors.New("no index selected (COLLSCAN)")
	}

	return qa.GetExplainDetails(doc), err
}

// GetExplainDetails returns summary from a doc
func (qa *QueryAnalyzer) GetExplainDetails(doc bson.M) ExplainSummary {
	summary := ExplainSummary{}
	// get shard name if a sharded cluster
	winningPlan := doc["queryPlanner"].(bson.M)["winningPlan"].(bson.M)
	if winningPlan["shards"] != nil {
		summary.ShardName = winningPlan["shards"].(primitive.A)[0].(bson.M)["shardName"].(string)
	}
	summary.ExecutionStats = getStageStats(doc["executionStats"].(bson.M))
	summary.AllPlansExecutionStats = []StageStats{}
	allPlansExecution := doc["executionStats"].(bson.M)["allPlansExecution"].(primitive.A)
	if len(allPlansExecution) > 0 && allPlansExecution[0].(bson.M)["allPlans"] != nil {
		allPlansExecution = allPlansExecution[0].(bson.M)["allPlans"].(primitive.A)
	}
	for _, execution := range allPlansExecution {
		summary.AllPlansExecutionStats = append(summary.AllPlansExecutionStats, getStageStats(execution.(bson.M)))
	}
	return summary
}

// GetSummary get summary of explain executionStats
func (qa *QueryAnalyzer) GetSummary(summary ExplainSummary) string {
	var buffer bytes.Buffer
	if summary.ShardName == "" {
		buffer.WriteString("Replica Set\n")
	} else {
		buffer.WriteString("Shard Name: " + summary.ShardName + "\n")
	}
	buffer.WriteString("\n=> Execution Stats\n")
	buffer.WriteString("=========================================\n")
	buffer.WriteString("Winning Plan:\n")
	buffer.WriteString(getStageStatsSummaryString(summary.ExecutionStats))

	buffer.WriteString("\n=> All Plans Execution\n")
	buffer.WriteString("=========================================\n")
	for i, stats := range summary.AllPlansExecutionStats {
		buffer.WriteString(fmt.Sprintf("Query Plan %d:\n", i+1))
		buffer.WriteString(getStageStatsSummaryString(stats))
		buffer.WriteString("\n")
	}
	return buffer.String()
}

// ReadQueryShapeFromFile gets filter map
func (qa *QueryAnalyzer) ReadQueryShapeFromFile(filename string) error {
	var err error
	var doc bson.M
	var buffer []byte
	if buffer, err = ioutil.ReadFile(filename); err != nil {
		return err
	}
	if err = json.Unmarshal(buffer, &doc); err == nil {
		if doc["filter"] != nil {
			qa.filter = doc["filter"].(map[string]interface{})
		}
		if doc["sort"] != nil {
			qa.sort = doc["sort"].(map[string]interface{})
		}
		return err
	}
	err = nil
	// can be a log entry
	re := regexp.MustCompile(`((\S+):)`)
	str := re.ReplaceAllString(string(buffer), "\"$2\":")
	ml := gox.NewMongoLog(str)
	filter := ml.Get(`"filter":`)
	re = regexp.MustCompile(`(new Date\(\S+\))`)
	filter = re.ReplaceAllString(filter, "\"$1\"")
	re = regexp.MustCompile(`ObjectId\(['"](\S+)['"]\)`)
	filter = re.ReplaceAllString(filter, "ObjectId('$1')")
	var f bson.M
	json.Unmarshal([]byte(filter), &f)
	d := gox.NewMapWalker(convert)
	qa.filter = d.Walk(f)
	sort := ml.Get(`"sort":`)
	json.Unmarshal([]byte(sort), &(qa.sort))
	return err
}

// https://github.com/mongodb/mongo/blob/master/src/mongo/db/query/plan_ranker.cpp
// we can run hint as {"explain": {"find": collectionName, "filter": filter, "sort": sortSpec, "hint": index}}
func getStageStats(execution bson.M) StageStats {
	summary := StageStats{TotalKeysExamined: execution["totalKeysExamined"].(int32),
		TotalDocsExamined: execution["totalDocsExamined"].(int32),
		InputStages:       []StageStats{}}
	executionStages := execution["executionStages"].(bson.M)
	inputStagesLevelArray := []inputStagesLevel{}
	if executionStages["inputStage"] != nil {
		inputStagesLevelArray = append(inputStagesLevelArray, getInputStagesLevel(executionStages["inputStage"].(bson.M), 0)...)
	} else if executionStages["inputStages"] != nil {
		inputStagesLevelArray = append(inputStagesLevelArray, inputStagesLevel{inputStages: executionStages["inputStages"].(primitive.A), level: 0})
	}
	stages := []string{}
	for _, elem := range inputStagesLevelArray {
		stages = append(stages, getAllStages(elem.inputStages)...)
	}
	advanced := executionStages["advanced"].(int32)
	works := executionStages["works"].(int32)
	summary.Score = getScore(advanced, works, stages)
	summary.Stage = executionStages["stage"].(string)
	b, _ := json.Marshal(executionStages["filter"])
	summary.Filter = gox.NewOrderedMap(string(b))
	summary.Advanced = advanced
	summary.Works = works
	summary.ExecTimeMillisEst = executionStages["executionTimeMillisEstimate"].(int32)

	for _, elem := range inputStagesLevelArray {
		stage := StageStats{Level: elem.level}
		for _, input := range elem.inputStages {
			inputStage := input.(bson.M)
			stage.Stage = inputStage["stage"].(string)
			if inputStage["keyPattern"] != nil {
				b, _ := json.Marshal(inputStage["keyPattern"])
				stage.KeyPattern = gox.NewOrderedMap(string(b))
			}
			if inputStage["filter"] != nil {
				b, _ := json.Marshal(inputStage["filter"])
				stage.Filter = gox.NewOrderedMap(string(b))
			}
			stage.Advanced = inputStage["advanced"].(int32)
			stage.Works = inputStage["works"].(int32)
			stage.ExecTimeMillisEst = inputStage["executionTimeMillisEstimate"].(int32)
			summary.InputStages = append(summary.InputStages, stage)
		}
	}
	return summary
}

func getStageStatsSummaryString(stat StageStats) string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("├─score: %v\n", stat.Score))
	buffer.WriteString(fmt.Sprintf("├─totalKeysExamined: %v\n", stat.TotalKeysExamined))
	buffer.WriteString(fmt.Sprintf("├─totalDocsExamined: %v\n", stat.TotalDocsExamined))
	buffer.WriteString(fmt.Sprintf("├─executionStages: %v\n", stat.Stage))
	if stat.Filter != nil {
		buffer.WriteString(fmt.Sprintf("├─  filter: %v\n", gox.Stringify(stat.Filter)))
	}
	buffer.WriteString(fmt.Sprintf("├─  advanced: %v\n", stat.Advanced))
	buffer.WriteString(fmt.Sprintf("├─  works: %v\n", stat.Works))
	buffer.WriteString(fmt.Sprintf("└─  executionTimeMillisEstimate: %v\n", stat.ExecTimeMillisEst))
	buffer.WriteString(getInputStagesSummaryString(stat.InputStages))
	return buffer.String()
}

func getInputStagesSummaryString(stats []StageStats) string {
	var buffer bytes.Buffer
	fillter := "  "
	for _, stat := range stats {
		sp := "  "
		for i := 0; i < stat.Level; i++ {
			sp += fillter
		}
		buffer.WriteString(fmt.Sprintf("%s├─inputStage: %v\n", sp, stat.Stage))
		if stat.KeyPattern != nil {
			buffer.WriteString(fmt.Sprintf("%s├─  key pattern: %v\n", sp, gox.Stringify(stat.KeyPattern)))
		}
		if stat.Filter != nil {
			buffer.WriteString(fmt.Sprintf("%s├─  filter: %v\n", sp, gox.Stringify(stat.Filter)))
		}
		buffer.WriteString(fmt.Sprintf("%s├─  advanced: %v\n", sp, stat.Advanced))
		buffer.WriteString(fmt.Sprintf("%s├─  works: %v\n", sp, stat.Works))
		buffer.WriteString(fmt.Sprintf("%s└─  executionTimeMillisEstimate: %v\n", sp, stat.ExecTimeMillisEst))
	}
	return buffer.String()
}

func getInputStagesLevel(inputStage bson.M, level int) []inputStagesLevel {
	inputStagesLevelArray := []inputStagesLevel{}
	inputStagesLevelArray = append(inputStagesLevelArray, inputStagesLevel{inputStages: primitive.A{inputStage}, level: level})
	if inputStage["inputStages"] != nil {
		inputStagesLevelArray = append(inputStagesLevelArray, inputStagesLevel{inputStages: inputStage["inputStages"].(primitive.A), level: (level + 1)})
	} else if inputStage["inputStage"] != nil {
		inputStagesLevelArray = append(inputStagesLevelArray, getInputStagesLevel(inputStage["inputStage"].(bson.M), level+1)...)
	}
	return inputStagesLevelArray
}

func getAllStages(inputStages primitive.A) []string {
	stages := []string{}
	for _, input := range inputStages {
		stages = append(stages, input.(bson.M)["stage"].(string))
	}
	return stages
}

// NewIndexScore returns index score
// scores refer to PlanRanker::scoreTree of plan_ranker.cpp
// score = baseScore + productivity + tieBreakers
// tieBreakers = noFetchBonus + noSortBonus + noIxisectBonus;
// noIxisectBonus: no STAGE_AND_HASH, STAGE_AND_SORTED
// noSortBonus: no STAGE_SORT
// by default noFetchBonus, noSortBonus, noIxisectBonus = epsilon
// epsilon = std::min(1.0 / static_cast<double>(10 * workUnits), 1e-4);
func getScore(advacned int32, works int32, stages []string) float64 {
	produtivity := float64(advacned) / float64(works)
	epsilon := math.Min(1/float64(works), .0001)
	noFetchBonus := epsilon
	if hasStage("FETCH", stages) &&
		(hasStage("PROJECTION_DEFAULT", stages) || hasStage("PROJECTION_COVERED", stages) || hasStage("PROJECTION_SIMPLE", stages)) {
		noFetchBonus = 0
	}
	noSortBonus := epsilon
	if hasStage("SORT", stages) {
		noSortBonus = 0
	}
	noIxisectBonus := epsilon
	if hasStage("AND_HASH", stages) || hasStage("AND_SORTED", stages) {
		noIxisectBonus = 0
	}
	tieBreakers := noFetchBonus + noSortBonus + noIxisectBonus
	score := 1 + produtivity + tieBreakers
	return score
}

func hasStage(stage string, stages []string) bool {
	for _, s := range stages {
		if s == stage {
			return true
		}
	}
	return false
}

func convert(x interface{}) interface{} {
	if v, ok := x.(string); ok {
		if strings.HasPrefix(v, "new Date(") {
			ms, _ := strconv.ParseInt(v[9:len(v)-1], 10, 64)
			return time.Unix(0, ms*int64(time.Millisecond))
		} else if strings.HasPrefix(v, "ObjectId(") {
			// expect ObjectId('<hex>')
			_id, _ := primitive.ObjectIDFromHex(v[10 : len(v)-2])
			return _id
		}
	}
	return x
}
