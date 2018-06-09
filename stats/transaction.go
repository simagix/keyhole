// Copyright 2018 Kuei-chun Chen. All rights reserved.

package stats

import (
	"encoding/json"
	"io/ioutil"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Transaction -
type Transaction struct {
	C string `json:"c"`
	Q bson.M `json:"q"`
	O bson.M `json:"o"`
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
		panic(err)
	}

	var doc TransactionDoc
	json.Unmarshal(bytes, &doc)
	return doc
}

func execTXForDemo(c *mgo.Collection, doc bson.M) int {
	schema := Schema{}
	results := []bson.M{}
	change := bson.M{"$set": bson.M{"ts": time.Now()}}
	bytes, _ := json.Marshal(doc)
	json.Unmarshal(bytes, &schema)
	city := schema.FavoriteCity
	book := schema.FavoriteBook
	movie := schema.FavoriteMovie
	txCount := 0

	c.Insert(doc)
	txCount++
	// c.Find(bson.M{"favoriteCity": city}).Sort("favoriteCity").Limit(512).All(&results)
	// txCount++
	c.Find(bson.M{"favoriteCity": city}).Limit(20).All(&results)
	txCount++
	c.Find(bson.M{"favoriteCity": city, "favoriteBook": book}).One(&results)
	txCount++
	c.Update(bson.M{"_id": doc["_id"]}, change)
	txCount++
	// c.Find(bson.M{"favoriteCity": city, "favoriteBook": book, "FavoriteMovie": movie}).One(&results)
	c.Find(bson.M{"favoritesList": bson.M{"$elemMatch": bson.M{"movie": movie}}}).One(&results)
	txCount++
	// c.Find(bson.M{"favoritesList": bson.M{"$elemMatch": bson.M{"book": book}}}).Limit(100).All(&results)
	// txCount++
	return txCount
}

func execTXByTemplateAndTX(c *mgo.Collection, doc bson.M, transactions []Transaction) int {
	results := []bson.M{}
	qfilter := make(map[string]interface{})
	var cmd = make(map[string]interface{})
	var op = make(map[string]interface{})

	for _, tx := range transactions {
		if tx.C == "insert" {
			c.Insert(doc)
		} else {
			bytes, _ := json.Marshal(tx.Q)
			json.Unmarshal(bytes, &cmd)
			traverseDocument(&qfilter, cmd, false)

			if tx.C == "find" {
				c.Find(qfilter).Limit(20).All(&results)
			} else if tx.C == "findOne" {
				c.Find(qfilter).One(&results)
			} else if tx.C == "update" {
				bytes, _ = json.Marshal(tx.O)
				json.Unmarshal(bytes, &op)
				traverseDocument(&qfilter, op, false)
				c.Update(qfilter, op)
			} else if tx.C == "updateAll" {
				bytes, _ = json.Marshal(tx.O)
				json.Unmarshal(bytes, &op)
				traverseDocument(&qfilter, op, false)
				c.UpdateAll(qfilter, op)
			} else if tx.C == "remove" {
				c.Remove(qfilter)
			} else if tx.C == "removeAll" {
				c.RemoveAll(qfilter)
			}
		}
	}

	return len(transactions)
}

func execTXByTemplate(c *mgo.Collection, doc bson.M) int {
	results := []bson.M{}
	change := bson.M{"$set": bson.M{"ts": time.Now()}}
	_id := doc["_id"]
	txCount := 0
	c.Insert(doc)
	txCount++
	c.Find(bson.M{"_id": _id}).One(&results)
	txCount++
	c.Update(bson.M{"_id": _id}, change)
	txCount++
	c.Find(bson.M{"_search": doc["_search"]}).Sort("_search").Limit(10).All(&results)
	txCount++
	c.Remove(bson.M{"_id": _id})
	txCount++
	return txCount
}
