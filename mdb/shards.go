// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"fmt"
	"net/url"
	"sort"
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
	Tags    []string       `bson:"tags"`
}

// GetShards return all shards from listShards command
func GetShards(client *mongo.Client) ([]Shard, error) {
	ctx := context.Background()
	var shardsInfo struct {
		Shards []Shard
	}
	if err := client.Database("admin").RunCommand(ctx, bson.D{{Key: "listShards", Value: 1}}).Decode(&shardsInfo); err != nil {
		return nil, err
	}
	sort.Slice(shardsInfo.Shards, func(i int, j int) bool {
		return shardsInfo.Shards[i].ID < shardsInfo.Shards[j].ID
	})
	return shardsInfo.Shards, nil
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
			ruri += connString.Username + ":" + url.QueryEscape(connString.Password) + "@" + hosts
		} else {
			ruri += hosts
		}
		ruri += fmt.Sprintf(`/%v?replicaSet=%v`, connString.Database, setName)
		if !isSRV && connString.AuthSource != "" {
			ruri += "&authSource=" + connString.AuthSource
		} else if isSRV {
			ruri += "&authSource=admin&tls=true"
		}
		ruri += GetQueryParams(connString, false)
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
				ruri += fmt.Sprintf(`%v:%v@%v/?connect=direct&`, connString.Username, url.QueryEscape(connString.Password), host)
			} else {
				ruri += fmt.Sprintf(`%v/?connect=direct&`, host)
			}
			if isSRV {
				ruri += "authSource=admin&tls=true"
			} else {
				if connString.AuthSource != "" {
					ruri += "authSource=" + connString.AuthSource
				} else if connString.Username != "" {
					ruri += "authSource=admin"
				}
			}
			ruri += GetQueryParams(connString, true)
			list = append(list, ruri)
		}
	}
	return list, nil
}

// GetQueryParams returns partial connection string from ConnString
func GetQueryParams(connString connstring.ConnString, isConnectDirect bool) string {
	ruri := ""
	if connString.SSLSet {
		ruri += "&tls=true"
	}
	if connString.SSLCaFileSet {
		ruri += "&tlsCAFile=" + connString.SSLCaFile
	}
	if connString.SSLClientCertificateKeyFileSet {
		ruri += "&tlsCertificateKeyFile=" + connString.SSLClientCertificateKeyFile
	}
	if connString.SSLInsecureSet {
		ruri += "&tlsInsecure=true"
	}
	if connString.ReadPreference != "" && !isConnectDirect {
		ruri += "&readPreference=" + connString.ReadPreference
	}
	if len(connString.ReadPreferenceTagSets) > 0 && !isConnectDirect {
		ruri += "&readPreferenceTags="
		cnt := 0
		for _, amap := range connString.ReadPreferenceTagSets {
			for k, v := range amap {
				ruri += k + ":" + v
				if cnt > 0 {
					ruri += ","
				}
				cnt++
			}
		}
	}
	if connString.WNumberSet {
		ruri += fmt.Sprintf("&w=%v", connString.WNumber)
	} else if connString.WString != "" {
		ruri += "&w=" + connString.WString
	}
	if connString.RetryReadsSet {
		ruri += "&retryReads=true"
	}
	if connString.RetryWritesSet {
		ruri += "&retryWrites=true"
	}
	return ruri
}
