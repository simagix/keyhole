// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"encoding/json"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

// ShardDoc information
type ShardDoc struct {
	ID    string `json:"_id" bson:"_id"`
	Host  string `json:"host" bson:"host"`
	State int    `json:"state" bson:"state"`
}

// GetShards gets a list of shards
func GetShards(client *mongo.Client) ([]ShardDoc, error) {
	var err error
	var m bson.M
	var buf []byte
	type shardInfo struct {
		Shards []ShardDoc `json:"shards"`
	}
	var si shardInfo
	if m, err = RunAdminCommand(client, "listShards"); err != nil {
		return nil, err
	}
	if buf, err = json.Marshal(m); err != nil {
		return nil, err
	}
	json.Unmarshal(buf, &si)
	return si.Shards, err
}

// GetShardListWithURI gets a list of shards
func GetShardListWithURI(client *mongo.Client, uri string) ([]string, error) {
	var err error
	var list []string

	isSRV := false
	if strings.Index(uri, "mongodb+srv") == 0 {
		isSRV = true
		uri = strings.Replace(uri, "mongodb+srv", "mongodb", 1)
	}
	var connString connstring.ConnString
	if connString, err = connstring.Parse(uri); err != nil {
		return list, err
	}

	var shards []ShardDoc
	if shards, err = GetShards(client); err != nil {
		return list, err
	}

	for s := range shards {
		host := shards[s].Host
		s := strings.Split(host, "/")
		if len(s) > 1 {
			connString.ReplicaSet = s[0]
		}
		ruri := "mongodb://"
		if connString.Username != "" {
			ruri += connString.Username + ":" + connString.Password + "@"
		}
		if len(s) == 1 {
			ruri += s[0] + "/?"
		} else {
			ruri += s[1] + "/?replicaSet=" + connString.ReplicaSet
		}
		if connString.AuthSource != "" {
			ruri += "&authSource=" + connString.AuthSource
		}
		if isSRV == true {
			ruri += "&authSource=admin&ssl=true"
		}
		list = append(list, ruri)
	}
	return list, nil
}
