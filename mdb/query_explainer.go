// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/simagix/gox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// QueryExplainer stores query analyzer info
type QueryExplainer struct {
	ExplainCmd ExplainCommand `bson:"explain"`
	NameSpace  string
	client     *mongo.Client
	document   bson.D
	isSharded  bool
	shardUsed  int
	verbose    bool
}

// ExplainCommand stores explain document
type ExplainCommand struct {
	Collection string `bson:"find"`
	Filter     bson.D `bson:"filter"`
	Sort       bson.D `bson:"sort,omitempty"`
	Hint       bson.D `bson:"hint,omitempty"`
	Group      string `bson:"group,omitempty"`
}

type inputStagesLevel struct {
	inputStages []bson.D
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

// IndexScore keeps index score
type IndexScore struct {
	Index gox.OrderedMap `json:"index"`
	Score float64        `json:"score"`
}

// NewQueryExplainer returns QueryExplainer
func NewQueryExplainer(client *mongo.Client) *QueryExplainer {
	return &QueryExplainer{client: client, ExplainCmd: ExplainCommand{}}
}

// SetVerbose sets verbosity
func (qe *QueryExplainer) SetVerbose(verbose bool) {
	qe.verbose = verbose
}

// GetFilter sets verbosity
func (qe *QueryExplainer) GetFilter() bson.D {
	return qe.ExplainCmd.Filter
}

// Explain explains query plans
func (qe *QueryExplainer) Explain() (ExplainSummary, error) {
	var err error
	var command bson.D
	o := QueryExplainer{}
	b, _ := bson.Marshal(qe)
	bson.Unmarshal(b, &o)
	o.ExplainCmd.Group = "" // remove group from index evaluation
	b, _ = bson.Marshal(o)
	bson.Unmarshal(b, &command)
	db := strings.Split(qe.NameSpace, ".")[0]
	if err = qe.client.Database(db).RunCommand(context.Background(), command).Decode(&qe.document); err != nil {
		return ExplainSummary{}, err
	}
	doc := qe.document.Map()
	winStage := doc["queryPlanner"].(bson.D).Map()["winningPlan"].(bson.D).Map()["stage"].(string)
	if winStage == "EOF" {
		return ExplainSummary{}, errors.New("no data found to be explained")
	} else if winStage == "COLLSCAN" {
		return ExplainSummary{}, errors.New("no index selected (COLLSCAN)")
	}

	return qe.GetExplainDetails(doc), err
}

// GetExplainDetails returns summary from a doc
func (qe *QueryExplainer) GetExplainDetails(doc bson.M) ExplainSummary {
	summary := ExplainSummary{}
	winningPlan := doc["queryPlanner"].(bson.D).Map()["winningPlan"].(bson.D).Map()
	if winningPlan["shards"] != nil {
		qe.isSharded = true
	}
	summary.ExecutionStats = qe.getStageStats(doc["executionStats"].(bson.D))
	summary.AllPlansExecutionStats = []StageStats{}

	allPlansExecution := doc["executionStats"].(bson.D).Map()["allPlansExecution"].(primitive.A)
	// pick a shard to evaluate if a sharded cluster
	if qe.isSharded == true && len(allPlansExecution) > 0 {
		maxReturned := int32(0)
		shardNames := []string{}
		for i, plansExecution := range allPlansExecution {
			shardNames = append(shardNames, plansExecution.(bson.D).Map()["shardName"].(string))
			allPlans := plansExecution.(bson.D).Map()["allPlans"].(primitive.A)
			for _, plan := range allPlans {
				rt := plan.(bson.D).Map()["nReturned"].(int32)
				if rt > maxReturned {
					maxReturned = rt
					qe.shardUsed = i
				}
			}
		}
		summary.ShardName = shardNames[qe.shardUsed]
		allPlansExecution = allPlansExecution[qe.shardUsed].(bson.D).Map()["allPlans"].(primitive.A)
	}
	for _, execution := range allPlansExecution {
		exec := execution.(bson.D)
		summary.AllPlansExecutionStats = append(summary.AllPlansExecutionStats, qe.getStageStats(exec))
	}
	return summary
}

// GetSummary get summary of explain executionStats
func (qe *QueryExplainer) GetSummary(summary ExplainSummary) string {
	var buffer bytes.Buffer
	buffer.WriteString("\n")
	if summary.ShardName == "" {
		buffer.WriteString("Cluster: Replica Set\n")
	} else {
		buffer.WriteString("Cluster: Sharded, evaluated from shard " + summary.ShardName + "\n")
	}
	b, _ := bson.Marshal(qe.ExplainCmd)
	var qshape bson.M
	bson.Unmarshal(b, &qshape)
	delete(qshape, "find")
	buffer.WriteString("Query Shape:\n" + gox.Stringify(qshape, "", "  ") + "\n")
	buffer.WriteString("\n=> Execution Stats\n")
	buffer.WriteString("=========================================\n")
	buffer.WriteString("Winning Plan:\n")
	buffer.WriteString(getStageStatsSummaryString(summary.ExecutionStats, 1))

	if len(summary.AllPlansExecutionStats) > 0 {
		buffer.WriteString("\n=> All Plans Execution\n")
		buffer.WriteString("=========================================\n")
		sort.Slice(summary.AllPlansExecutionStats, func(i, j int) bool {
			return summary.AllPlansExecutionStats[i].Score > summary.AllPlansExecutionStats[j].Score
		})
		for i, stats := range summary.AllPlansExecutionStats {
			buffer.WriteString(fmt.Sprintf("Query Plan %d:\n", i+1))
			buffer.WriteString(getStageStatsSummaryString(stats, 0))
			buffer.WriteString("\n")
		}
	}
	return buffer.String()
}

func (qe *QueryExplainer) getDocument(depth []string) interface{} {
	doc := qe.document
	for _, key := range depth {
		for _, value := range doc {
			if key == value.Key {
				doc = value.Value.(bson.D)
			}
		}
	}
	fmt.Println(gox.Stringify(doc))
	return doc
}

// https://github.com/mongodb/mongo/blob/master/src/mongo/db/query/plan_ranker.cpp
// we can run hint as {"explain": {"find": collectionName, "filter": filter, "sort": sortSpec, "hint": index}}
func (qe *QueryExplainer) getStageStats(document bson.D) StageStats {
	execution := document.Map()
	summary := StageStats{TotalKeysExamined: execution["totalKeysExamined"].(int32),
		TotalDocsExamined: execution["totalDocsExamined"].(int32),
		InputStages:       []StageStats{}}
	executionStages := execution["executionStages"].(bson.D).Map()
	if qe.isSharded {
		shard := StageStats{TotalKeysExamined: execution["totalKeysExamined"].(int32),
			TotalDocsExamined: execution["totalDocsExamined"].(int32),
			InputStages:       []StageStats{}}
		shard.Stage = executionStages["stage"].(string)
		shards, ok := executionStages["shards"].(primitive.A)
		if ok {
			executionStages = shards[qe.shardUsed].(bson.D).Map()["executionStages"].(bson.D).Map()
		}
	}
	inputStagesLevelArray := []inputStagesLevel{}
	if executionStages["inputStage"] != nil {
		inputStagesLevelArray = append(inputStagesLevelArray, getInputStagesLevel(executionStages["inputStage"].(bson.D), 0)...)
	} else if executionStages["inputStages"] != nil {
		for _, istage := range executionStages["inputStages"].(primitive.A) {
			inputStagesLevelArray = append(inputStagesLevelArray, getInputStagesLevel(istage.(bson.D), 0)...)
		}
	}
	stages := []string{}
	for _, elem := range inputStagesLevelArray {
		stages = append(stages, getAllStages(elem.inputStages)...)
	}
	advanced := executionStages["advanced"].(int32)
	works := executionStages["works"].(int32)
	summary.Score = getScore(advanced, works, stages)
	summary.Stage = executionStages["stage"].(string)
	if executionStages["filter"] != nil {
		var v bson.M
		b, _ := bson.Marshal(executionStages["filter"].(bson.D))
		bson.Unmarshal(b, &v)
		b, _ = json.Marshal(v)
		summary.Filter = gox.NewOrderedMap(string(b))
	}
	summary.Advanced = advanced
	summary.Works = works
	summary.ExecTimeMillisEst = executionStages["executionTimeMillisEstimate"].(int32)
	for _, elem := range inputStagesLevelArray {
		stage := StageStats{Level: elem.level}
		for _, input := range elem.inputStages {
			inputStage := input.Map()
			stage.Stage = inputStage["stage"].(string)
			if inputStage["keyPattern"] != nil {
				str := "{"
				for i, v := range inputStage["keyPattern"].(bson.D) {
					if i > 0 {
						str += ","
					}
					str += fmt.Sprintf(`"%v":%v`, v.Key, v.Value)
				}
				str += "}"
				stage.KeyPattern = gox.NewOrderedMap(str)
			}
			if inputStage["filter"] != nil {
				var v bson.M
				b, _ := bson.Marshal(inputStage["filter"].(bson.D))
				bson.Unmarshal(b, &v)
				b, _ = json.Marshal(v)
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

// GetIndexesScores returns a list of indexes scores
func (qe *QueryExplainer) GetIndexesScores(keys []string) []IndexScore {
	var err error
	var cur *mongo.Cursor
	var indexes []string
	scores := []IndexScore{}
	ctx := context.Background()
	pos := strings.Index(qe.NameSpace, ".")
	db := qe.NameSpace[:pos]
	coll := qe.NameSpace[pos+1:]
	collection := qe.client.Database(db).Collection(coll)
	indexView := collection.Indexes()
	if cur, err = indexView.List(ctx); err != nil {
		log.Println(err)
		return scores
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var idx = bson.D{}
		if err = cur.Decode(&idx); err != nil {
			continue
		}

		var keys bson.D
		for _, v := range idx {
			if v.Key == "key" {
				keys = v.Value.(bson.D)
				break
			}
		}
		var strbuf bytes.Buffer
		fields := []string{}
		for n, value := range keys {
			fields = append(fields, value.Key)
			if n == 0 {
				strbuf.WriteString("{")
			}

			vt := reflect.TypeOf(value.Value)
			switch vt.Kind() {
			case reflect.String:
				strbuf.WriteString(fmt.Sprintf(`"%v":"%v"`, value.Key, value.Value))
			default:
				strbuf.WriteString(fmt.Sprintf(`"%v":%v`, value.Key, value.Value))
			}
			if n == len(keys)-1 {
				strbuf.WriteString("}")
			} else {
				strbuf.WriteString(",")
			}
		}
		indexes = append(indexes, strbuf.String())
	}
	keyMap := make(map[string]string)
	for i := 0; i < len(keys); i++ {
		keyMap[keys[i]] = "v"
	}
	// Execute explain on all indexes
	for _, index := range indexes {
		bson.UnmarshalExtJSON([]byte(index), true, &qe.ExplainCmd.Hint)
		o := QueryExplainer{}
		b, _ := bson.Marshal(qe)
		bson.Unmarshal(b, &o)
		o.ExplainCmd.Group = "" // remove group from index evaluation
		b, _ = bson.Marshal(o)
		var cmd bson.D
		bson.Unmarshal(b, &cmd)
		filter := cmd.Map()["explain"].(bson.D).Map()["hint"].(bson.D)
		if len(filter) == 0 || keyMap[filter[0].Key] == "" {
			continue
		}
		var document = bson.D{}
		if err = collection.Database().RunCommand(ctx, cmd).Decode(&document); err != nil {
			fmt.Println(err.Error())
			continue
		}
		summary := qe.GetExplainDetails(document.Map())
		stages := []string{}
		for _, elem := range summary.ExecutionStats.InputStages {
			stages = append(stages, elem.Stage)
		}

		score := getScore(summary.ExecutionStats.Advanced, summary.ExecutionStats.Works, stages)
		om := gox.NewOrderedMap(index)
		scores = append(scores, IndexScore{Index: *om, Score: score})
	}

	// sorted by score DESC
	sort.Slice(scores, func(i, j int) bool {
		if scores[i].Score > scores[j].Score {
			return true
		} else if scores[i].Score == scores[j].Score {
			return gox.Stringify(scores[i].Index) < gox.Stringify(scores[j].Index)
		}
		return false
	})
	return scores
}

// ReadQueryShape parses filter map
func (qe *QueryExplainer) ReadQueryShape(buffer []byte) error {
	var err error
	var doc bson.D
	var ns string
	explainCmd := ExplainCommand{}
	if err = bson.UnmarshalExtJSON(buffer, true, &doc); err == nil {
		if doc.Map()["filter"] != nil {
			explainCmd.Filter = doc.Map()["filter"].(bson.D)
		}
		if doc.Map()["sort"] != nil {
			explainCmd.Sort = doc.Map()["sort"].(bson.D)
		}
		if doc.Map()["hint"] != nil {
			explainCmd.Hint = doc.Map()["hint"].(bson.D)
		}
		ns = doc.Map()["ns"].(string)
		pos := strings.Index(ns, ".")
		explainCmd.Collection = ns[pos+1:]
		return err
	}
	err = nil
	// can be a log entry
	re := regexp.MustCompile(`((\S+):)`)
	str := re.ReplaceAllString(string(buffer), "\"$2\":")
	ml := gox.NewMongoLog(str)
	filter := ml.Get(`"filter":`)
	group := ""
	if filter == "" {
		filter = ml.Get(`"$match":`)
		if filter != "" {
			group = ml.Get(`"$group":`)
		}
	}
	if filter == "" {
		filter = ml.Get(`"query":`)
	}
	if group != "" {
		d := bson.M{}
		bson.UnmarshalExtJSON([]byte(group), true, &d)
		if d["_id"] != "" {
			explainCmd.Group = d["_id"].(string)[1:]
		}
	}
	re = regexp.MustCompile(`(new Date\(\S+\))`)
	filter = re.ReplaceAllString(filter, "\"$1\"")
	re = regexp.MustCompile(`ObjectId\(['"](\S+)['"]\)`)
	filter = re.ReplaceAllString(filter, "ObjectId('$1')")
	re = regexp.MustCompile(`\/(\S+)\/(\S+)?`)
	filter = re.ReplaceAllString(filter, "{ \"$$regex\": \"$1\", \"$$options\": \"$2\" }")
	var f bson.M
	json.Unmarshal([]byte(filter), &f)
	d := gox.NewMapWalker(convert)
	docMap := d.Walk(f)
	b, _ := bson.Marshal(docMap)
	bson.Unmarshal(b, &explainCmd.Filter)
	sort := ml.Get(`"sort":`)
	if sort == "" {
		sort = ml.Get(`"$sort":`)
	}
	bson.UnmarshalExtJSON([]byte(sort), true, &(explainCmd.Sort))
	xs := string(buffer)
	i := strings.Index(xs, "] ")
	ns = strings.Split(xs[i+2:], " ")[1]
	pos := strings.Index(ns, ".")
	explainCmd.Collection = ns[pos+1:]
	qe.ExplainCmd = explainCmd
	qe.NameSpace = ns
	return err
}

func getStageStatsSummaryString(stat StageStats, level int) string {
	var buffer bytes.Buffer
	if stat.Stage == "SHARD_MERGE" {
		buffer.WriteString(fmt.Sprintf("└─executionStages: %v\n", stat.Stage))
		return buffer.String()
	}
	sp := ""
	for i := 0; i < level; i++ {
		sp += "  "
	}
	buffer.WriteString(fmt.Sprintf("%s├─score: %v\n", sp, stat.Score))
	buffer.WriteString(fmt.Sprintf("%s├─totalKeysExamined: %v\n", sp, stat.TotalKeysExamined))
	buffer.WriteString(fmt.Sprintf("%s├─totalDocsExamined: %v\n", sp, stat.TotalDocsExamined))
	buffer.WriteString(fmt.Sprintf("%s├─executionStages: %v\n", sp, stat.Stage))
	if stat.Filter != nil {
		buffer.WriteString(fmt.Sprintf("%s├─  filter: %v\n", sp, gox.Stringify(stat.Filter)))
	}
	buffer.WriteString(fmt.Sprintf("%s├─  advanced: %v\n", sp, stat.Advanced))
	buffer.WriteString(fmt.Sprintf("%s├─  works: %v\n", sp, stat.Works))
	buffer.WriteString(fmt.Sprintf("%s└─  executionTimeMillisEstimate: %v\n", sp, stat.ExecTimeMillisEst))
	buffer.WriteString(getInputStagesSummaryString(stat.InputStages, level))
	return buffer.String()
}

func getInputStagesSummaryString(stats []StageStats, level int) string {
	var buffer bytes.Buffer
	fillter := "  "
	for _, stat := range stats {
		sp := "  "
		for i := 0; i < stat.Level+level; i++ {
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

func getInputStagesLevel(inputStage bson.D, level int) []inputStagesLevel {
	inputStagesLevelArray := []inputStagesLevel{}
	inputStagesLevelArray = append(inputStagesLevelArray, inputStagesLevel{inputStages: []bson.D{inputStage}, level: level})
	if inputStage.Map()["inputStages"] != nil {
		for _, istage := range inputStage.Map()["inputStages"].(primitive.A) {
			inputStagesLevelArray = append(inputStagesLevelArray, getInputStagesLevel(istage.(bson.D), level+1)...)
		}
	} else if inputStage.Map()["inputStage"] != nil {
		inputStagesLevelArray = append(inputStagesLevelArray, getInputStagesLevel(inputStage.Map()["inputStage"].(bson.D), level+1)...)
	}
	return inputStagesLevelArray
}

func getAllStages(inputStages []bson.D) []string {
	stages := []string{}
	for _, input := range inputStages {
		stages = append(stages, input.Map()["stage"].(string))
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
