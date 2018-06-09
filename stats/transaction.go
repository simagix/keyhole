package stats

import (
	"encoding/json"
	"io/ioutil"

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
func (m MongoConn) GetTransactions(filename string) TransactionDoc {
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
