// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"fmt"
	"html/template"
	"strings"

	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Shard store shard information
type Shard struct {
	ID      string         `bson:"_id"`
	Host    string         `bson:"host"`
	State   int            `bson:"state"`
	Servers []ClusterStats `bson:"servers"`
}

// GetShards return all shards from listShards command
func GetShards(client *mongo.Client) ([]Shard, error) {
	ctx := context.Background()
	var shardsInfo struct {
		Shards []Shard
	}
	err := client.Database("admin").RunCommand(ctx, bson.D{{Key: "listShards", Value: 1}}).Decode(&shardsInfo)
	return shardsInfo.Shards, err
}

// GetAllShardURIs returns URIs of all replicas
func GetAllShardURIs(shards []Shard, connString connstring.ConnString) ([]string, error) {
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
			ruri += connString.Username + ":" + template.URLQueryEscaper(connString.Password) + "@" + hosts
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
func GetAllServerURIs(shards []Shard, connString connstring.ConnString) ([]string, error) {
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
			if isSRV == true {
				ruri += "authSource=admin&ssl=true"
			} else {
				if connString.AuthSource != "" {
					ruri += "authSource=" + connString.AuthSource
				} else if connString.Username != "" {
					ruri += "authSource=admin"
				}
				if connString.SSLCaFile != "" {
					ruri += "&ssl=true"
				}
			}
			list = append(list, ruri)
		}
	}
	return list, nil
}
