package stats

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

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
	Transactions []Transaction
	Indexes      []bson.M
}

// GetTransactions -
func GetTransactions(filename string) TransactionDoc {
	if filename == "" {
		return TransactionDoc{}
	}
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var doc TransactionDoc
	json.Unmarshal(bytes, &doc)
	return doc
}

func (m MongoConn) processTransactions(transactions []Transaction, c *mgo.Collection, doc bson.M) int {
	results := []bson.M{}
	qfilter := make(map[string]interface{})
	var cmd = make(map[string]interface{})
	var op = make(map[string]interface{})

	for _, tx := range transactions {
		if tx.C == "insert" {
			c.Insert(cloneDoc(doc))
		} else {
			bytes, _ := json.Marshal(tx.Q)
			json.Unmarshal(bytes, &cmd)
			traverseDocument(&qfilter, cmd, false)

			if tx.C == "find" {
				c.Find(qfilter).All(&results)
			} else if tx.C == "update" {
				bytes, _ = json.Marshal(tx.O)
				json.Unmarshal(bytes, &op)
				traverseDocument(&qfilter, op, false)
				c.Update(qfilter, op)
			} else if tx.C == "remove" {
				c.Remove(qfilter)
			}
		}
	}

	return len(transactions)
}
