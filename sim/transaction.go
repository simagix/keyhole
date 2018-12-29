// Copyright 2018 Kuei-chun Chen. All rights reserved.

package sim

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
	"github.com/simagix/keyhole/mdb"
	"github.com/simagix/keyhole/sim/util"
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

func execTXForDemo(c *mongo.Collection, doc bson.M) int {
	var err error
	ctx := context.Background()
	schema := util.FavoritesSchema{}
	change := bson.M{"$set": bson.M{"ts": time.Now()}}
	bytes, _ := json.Marshal(doc)
	json.Unmarshal(bytes, &schema)
	city := schema.FavoriteCity
	book := schema.FavoriteBook
	movie := schema.FavoriteMovie
	txCount := 0

	if _, err = c.InsertOne(ctx, doc); err != nil {
		panic(err)
	}
	txCount++
	// c.Find(bson.M{"favoriteCity": city}).Sort("favoriteCity").Limit(512).All(&results)
	// txCount++
	opts := options.Find()
	opts.SetLimit(20)
	c.Find(ctx, bson.D{{Key: "favoriteCity", Value: city}}, opts)
	txCount++
	c.Find(ctx, bson.D{{Key: "favoriteCity", Value: city}, {Key: "favoriteBook", Value: book}}, opts)
	txCount++
	c.UpdateOne(ctx, bson.M{"_id": doc["_id"]}, change)
	txCount++
	// c.Find(bson.M{"favoriteCity": city, "favoriteBook": book, "FavoriteMovie": movie}).One(&results)
	opts.SetSort(bson.D{{Key: "favoriteCity", Value: 1}})
	c.Find(ctx, bson.M{"favoritesList": bson.M{"$elemMatch": bson.M{"movie": movie}}}, opts)
	txCount++
	// c.Find(bson.M{"favoritesList": bson.M{"$elemMatch": bson.M{"book": book}}}).Limit(100).All(&results)
	// txCount++
	return txCount
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
				// var pipeline []bson.D
				// bytes, _ := json.Marshal(tx.Pipe)
				// json.Unmarshal(bytes, &pipeline)
				// example
				// var pipeline = mongo.Pipeline{{
				// 	{"$group", bson.D{{"_id", "$state"}, {"totalPop", bson.D{{"$sum", "$pop"}}}}},
				// 	{"$match", bson.D{{"totalPop", bson.D{{"$gte", 10 * 1000 * 1000}}}}},
				// }}
				b, _ := json.Marshal(tx.Pipe)
				c.Aggregate(ctx, mdb.MongoPipeline(string(b)))
			}
		}
	}

	return len(transactions)
}

func execTXByTemplate(c *mongo.Collection, doc bson.M) int {
	ctx := context.Background()
	change := bson.M{"$set": bson.M{"ts": time.Now()}}
	_id := doc["_id"]
	txCount := 0
	c.InsertOne(ctx, doc)
	txCount++
	c.FindOne(ctx, bson.M{"_id": _id})
	txCount++
	c.UpdateOne(ctx, bson.M{"_id": _id}, change)
	txCount++
	opts := options.Find()
	opts.SetLimit(20)
	c.Find(ctx, bson.M{"_search": doc["_search"]}, opts)
	txCount++
	c.DeleteOne(ctx, bson.M{"_id": _id})
	txCount++
	return txCount
}
