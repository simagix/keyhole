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
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func execTXByTemplateAndTX(c *mongo.Collection, doc bson.M, transactions []Transaction) (bson.M, error) {
	var err error
	var ctx = context.Background()
	var op = make(map[string]interface{})
	var execTime = bson.M{}

	for _, tx := range transactions {
		t := time.Now()
		if tx.C == "insertOne" {
			if _, err = c.InsertOne(ctx, doc); err != nil {
				return execTime, err
			}
		} else {
			bytes, _ := json.Marshal(tx.Filter)
			cmd := make(map[string]interface{})
			filter := make(map[string]interface{})
			json.Unmarshal(bytes, &cmd)
			util.RandomizeDocument(&filter, cmd, false)

			if tx.C == "find" {
				if _, err = c.Find(ctx, filter); err != nil {
					return execTime, err
				}
			} else if tx.C == "findOne" {
				if r := c.FindOne(ctx, filter); r.Err() != nil {
					return execTime, err
				}
			} else if tx.C == "updateOne" {
				bytes, _ = json.Marshal(tx.Op)
				json.Unmarshal(bytes, &op)
				util.RandomizeDocument(&filter, op, false)
				if _, err = c.UpdateOne(ctx, filter, op); err != nil {
					return execTime, err
				}
			} else if tx.C == "updateMany" {
				bytes, _ = json.Marshal(tx.Op)
				json.Unmarshal(bytes, &op)
				util.RandomizeDocument(&filter, op, false)
				if _, err = c.UpdateMany(ctx, filter, op); err != nil {
					return execTime, err
				}
			} else if tx.C == "deleteOne" {
				if _, err = c.DeleteOne(ctx, filter); err != nil {
					return execTime, err
				}
			} else if tx.C == "deleteMany" {
				if _, err = c.DeleteMany(ctx, filter); err != nil {
					return execTime, err
				}
			} else if tx.C == "aggregate" {
				// example
				// pipeline := mongo.Pipeline{
				// 	{{"$group", bson.D{{"_id", "$state"}, {"totalPop", bson.D{{"$sum", "$pop"}}}}}},
				// 	{{"$match", bson.D{{"totalPop", bson.D{{"$gte", 10 * 1000 * 1000}}}}}},
				// }
				b, _ := json.Marshal(tx.Pipe)
				if _, err = c.Aggregate(ctx, mdb.MongoPipeline(string(b))); err != nil {
					return execTime, err
				}
			}
			execTime[tx.C] = time.Now().Sub(t)
		}
	}

	execTime["total"] = len(transactions)
	return execTime, err
}

func execTx(c *mongo.Collection, doc bson.M) (bson.M, error) {
	var err error
	var tm []time.Time
	var execTime = bson.M{}
	ctx := context.Background()
	ts := time.Now()
	change := bson.M{"$set": bson.M{"timestamp": ts}}
	o := primitive.NewObjectID()
	doc["_id"] = o
	doc["ts"] = ts
	tm = append(tm, time.Now())
	if _, err = c.InsertOne(ctx, doc); err != nil {
		return execTime, err
	}
	filter := bson.D{{Key: "_id", Value: o}}
	tm = append(tm, time.Now())
	if c.FindOne(ctx, filter).Err() != nil {
		return execTime, err
	}
	tm = append(tm, time.Now())
	if _, err = c.UpdateOne(ctx, filter, change); err != nil {
		return execTime, err
	}
	tm = append(tm, time.Now())
	if _, err = c.DeleteOne(ctx, filter); err != nil {
		return execTime, err
	}
	tm = append(tm, time.Now())
	keys := []string{"InsertOne", "FindOne", "UpdateOne", "DeleteOne"}
	for i := 1; i < len(tm); i++ {
		execTime[keys[i-1]] = tm[i].Sub(tm[i-1])
	}
	execTime["total"] = len(keys)
	return execTime, err
}
