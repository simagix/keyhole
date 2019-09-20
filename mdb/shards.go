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

// ShardList -
type ShardList struct {
	Shards []ShardDoc `json:"shards" bson:"shards"`
	OK     int        `json:"ok" bson:"ok"`
}

// GetShards -
func GetShards(client *mongo.Client, uri string) ([]string, error) {
	var err error
	var m bson.M
	var list []string

	if m, err = RunAdminCommand(client, "listShards"); err != nil {
		return list, err
	}

	isSRV := false
	if strings.Index(uri, "mongodb+srv") == 0 {
		isSRV = true
		uri = strings.Replace(uri, "mongodb+srv", "mongodb", 1)
	}
	var connString connstring.ConnString
	if connString, err = connstring.Parse(uri); err != nil {
		panic(err)
	}
	shards := ShardList{}
	bytes, _ := json.Marshal(m)
	json.Unmarshal(bytes, &shards)

	for s := range shards.Shards {
		host := shards.Shards[s].Host
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

// GetShardsURIList gets shards list
func GetShardsURIList(client *mongo.Client, uri string) ([]string, error) {
	var uriList []string
	var err error
	var ssi ServerInfo
	if ssi, err = GetServerInfo(client); err != nil {
		return uriList, err
	}
	uriList = []string{uri}
	if ssi.Cluster == SHARDED {
		if uriList, err = GetShards(client, uri); err != nil {
			return uriList, err
		}
	}
	return uriList, err
}
