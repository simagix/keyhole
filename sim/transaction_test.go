// Copyright 2018 Kuei-chun Chen. All rights reserved.

package sim

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/simagix/keyhole/mdb"
	"github.com/simagix/keyhole/sim/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestGetTransactions(t *testing.T) {
	TransactionDoc := GetTransactions("")
	bytes, _ := json.MarshalIndent(TransactionDoc, "", "  ")
	t.Log(string(bytes))
}

func TestExecTXByTemplateAndTX(t *testing.T) {
	var filename = "testdata/transactions.json"
	var client *mongo.Client
	client = getMongoClient()
	defer client.Disconnect(context.Background())
	c := client.Database(mdb.KEYHOLEDB).Collection("examples")
	tx := GetTransactions(filename)
	_, err := execTXByTemplateAndTX(c, util.GetDemoDoc(), tx.Transactions[0])
	if err != nil {
		t.Fatal(err)
	}
}

func TestExecTx(t *testing.T) {
	var err error
	var client *mongo.Client
	var tm bson.M
	client = getMongoClient()
	defer client.Disconnect(context.Background())
	c := client.Database(mdb.KEYHOLEDB).Collection("examples")
	if tm, err = execTx(c, util.GetDemoDoc()); err != nil {
		t.Fatal()
	}

	for k, v := range tm {
		t.Log(k, v)
	}
}
