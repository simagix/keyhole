// Copyright 2018 Kuei-chun Chen. All rights reserved.

package sim

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/simagix/keyhole/mdb"
	"github.com/simagix/keyhole/sim/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Transaction -
type Transaction struct {
	C      string   `json:"c"`
	Filter bson.M   `json:"filter"`
	Op     bson.M   `json:"op"`
	Pipe   []bson.M `json:"pipe"`
}

// TransactionDoc -
type TransactionDoc struct {
	Transactions []Transaction `json:"transactions" bson:"transactions"`
	Indexes      []bson.M      `json:"indexes" bson:"indexes"`
}

// GetTransactions -
func GetTransactions(filename string) TransactionDoc {
	if filename == "" {
		return TransactionDoc{}
	}
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return TransactionDoc{}
	}

	var doc TransactionDoc
	json.Unmarshal(bytes, &doc)
	return doc
}

func execTXByTemplateAndTX(c *mongo.Collection, doc bson.M, transactions []Transaction) int {
	var ctx = context.Background()
	var op = make(map[string]interface{})

	for _, tx := range transactions {
		if tx.C == "insert" {
			c.InsertOne(ctx, doc)
		} else {
			bytes, _ := json.Marshal(tx.Filter)
			cmd := make(map[string]interface{})
			filter := make(map[string]interface{})
			json.Unmarshal(bytes, &cmd)
			util.RandomizeDocument(&filter, cmd, false)

			if tx.C == "find" {
				c.Find(ctx, filter)
			} else if tx.C == "findOne" {
				c.FindOne(ctx, filter)
			} else if tx.C == "update" {
				bytes, _ = json.Marshal(tx.Op)
				json.Unmarshal(bytes, &op)
				util.RandomizeDocument(&filter, op, false)
				c.UpdateMany(ctx, filter, op)
			} else if tx.C == "updateAll" || tx.C == "updateMany" {
				bytes, _ = json.Marshal(tx.Op)
				json.Unmarshal(bytes, &op)
				util.RandomizeDocument(&filter, op, false)
				c.UpdateMany(ctx, filter, op)
			} else if tx.C == "remove" || tx.C == "deleteOne" {
				c.DeleteOne(ctx, filter)
			} else if tx.C == "removeAll" || tx.C == "deleteMany" {
				c.DeleteMany(ctx, filter)
			} else if tx.C == "aggregate" {
				// example
				// pipeline := mongo.Pipeline{
				// 	{{"$group", bson.D{{"_id", "$state"}, {"totalPop", bson.D{{"$sum", "$pop"}}}}}},
				// 	{{"$match", bson.D{{"totalPop", bson.D{{"$gte", 10 * 1000 * 1000}}}}}},
				// }
				b, _ := json.Marshal(tx.Pipe)
				c.Aggregate(ctx, mdb.MongoPipeline(string(b)))
			}
		}
	}

	return len(transactions)
}

func execTx(c *mongo.Collection, doc bson.M) (bson.M, error) {
	var err error
	var results *mongo.InsertManyResult
	var docs []interface{}
	var tm []time.Time
	var execTime = bson.M{}
	ctx := context.Background()
	ts := time.Now()
	change := bson.M{"$set": bson.M{"timestamp": ts}}

	for i := 0; i < 3; i++ {
		d := util.CloneDoc(doc)
		d["ts"] = ts
		delete(d, "_id")
		docs = append(docs, d)
	}
	tm = append(tm, time.Now())
	if results, err = c.InsertMany(ctx, docs); err != nil {
		return execTime, err
	}
	filters := bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: results.InsertedIDs}}}}
	filter := bson.D{{Key: "_id", Value: results.InsertedIDs[0]}}
	tm = append(tm, time.Now())
	if c.FindOne(ctx, filter).Err() != nil {
		return execTime, err
	}
	tm = append(tm, time.Now())
	if _, err = c.Find(ctx, filters); err != nil {
		return execTime, err
	}
	tm = append(tm, time.Now())
	if _, err = c.UpdateOne(ctx, filter, change); err != nil {
		return execTime, err
	}
	tm = append(tm, time.Now())
	if _, err = c.UpdateMany(ctx, filters, change); err != nil {
		return execTime, err
	}
	tm = append(tm, time.Now())
	if _, err = c.DeleteOne(ctx, filter); err != nil {
		return execTime, err
	}
	tm = append(tm, time.Now())
	if _, err = c.DeleteMany(ctx, filters); err != nil {
		return execTime, err
	}
	tm = append(tm, time.Now())
	keys := []string{"InsertMany", "FindOne", "Find", "UpdateOne", "UpdateMany", "DeleteOne", "DeleteMany"}
	for i := 1; i < len(tm); i++ {
		execTime[keys[i-1]] = tm[i].Sub(tm[i-1])
	}
	return execTime, err
}
