// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"encoding/json"
	"fmt"
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

// GetAllShardURIs returns URIs of all replicas
func GetAllShardURIs(shards []ShardDoc, connString connstring.ConnString) ([]string, error) {
	var list []string
	isSRV := false
	if strings.Index(connString.String(), "mongodb+srv") == 0 {
		isSRV = true
	}
	for _, shard := range shards {
		idx := strings.Index(shard.Host, "/")
		setName := shard.Host[:idx]
		hosts := shard.Host[idx+1:]
		ruri := "mongodb://"
		if connString.Username != "" {
			ruri += connString.Username + ":" + connString.Password + "@" + hosts
		} else {
			ruri += hosts
		}
		ruri += fmt.Sprintf(`/%v?replicaSet=%v`, connString.Database, setName)
		if isSRV == false && connString.AuthSource != "" {
			ruri += "&authSource=" + connString.AuthSource
		} else if isSRV == true {
			ruri += "&authSource=admin&ssl=true"
		}
		list = append(list, ruri)
	}
	return list, nil
}

// GetAllServerURIs returns URIs of all mongo servers
func GetAllServerURIs(shards []ShardDoc, connString connstring.ConnString) ([]string, error) {
	var list []string
	isSRV := false
	if strings.HasPrefix(connString.String(), "mongodb+srv") {
		isSRV = true
	}
	for _, shard := range shards {
		idx := strings.Index(shard.Host, "/")
		hosts := strings.Split(shard.Host[idx+1:], ",")
		for _, host := range hosts {
			ruri := "mongodb://"
			if connString.Username != "" {
				ruri += fmt.Sprintf(`%v:%v@%v/?`, connString.Username, connString.Password, host)
			} else {
				ruri += fmt.Sprintf(`%v/?`, host)
			}
			if isSRV == false && connString.AuthSource != "" {
				ruri += "authSource=" + connString.AuthSource
			} else if isSRV == true {
				ruri += "authSource=admin&ssl=true"
			}
			list = append(list, ruri)
		}
	}
	return list, nil
}
