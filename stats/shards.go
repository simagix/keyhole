// Copyright 2018 Kuei-chun Chen. All rights reserved.

package stats

import (
	"encoding/json"
	"fmt"
	"strings"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// ShardDoc -
type Shard struct {
	ID    string `json:"_id" bson:"_id"`
	Host  string `json:"host" bson:"host"`
	State int    `json:"state" bson:"state"`
}

// ShardList -
type ShardList struct {
	Shards []Shard `json:"shards" bson:"shards"`
	OK     int     `json:"ok" bson:"ok"`
}

// GetShards -
func GetShards(session *mgo.Session, uri string) []string {
	var list []string
	result := bson.M{}
	if err := session.DB("admin").Run("listShards", &result); err != nil {
		panic(err)
	}

	dialInfo, _ := mgo.ParseURL(uri)
	shards := ShardList{}
	bytes, _ := json.Marshal(result)
	json.Unmarshal(bytes, &shards)

	for s := range shards.Shards {
		host := shards.Shards[s].Host

		s := strings.Split(host, "/")
		dialInfo.ReplicaSetName = s[0]
		s = strings.Split(s[1], ",")
		dialInfo.Addrs = s
		ruri := "mongodb://"
		if dialInfo.Username != "" {
			ruri = ruri + dialInfo.Username + ":" + dialInfo.Password + "@"
		}
		ruri = ruri + s[0] + "?replicaSet=" + dialInfo.ReplicaSetName
		if dialInfo.Source != "" {
			ruri = ruri + "&authSource=" + dialInfo.Source
		}

		list = append(list, ruri)
	}

	return list
}

// ShardCollection -
func ShardCollection(session *mgo.Session, collname string) {
	fmt.Println("Shard", collname)
	result := bson.M{}
	if err := session.DB("admin").Run(bson.D{{Name: "enableSharding", Value: DBName}}, &result); err != nil {
		fmt.Println(err)
	}
	if err := session.DB("admin").Run(bson.D{{Name: "shardCollection", Value: collname}, {Name: "key", Value: bson.M{"_id": "hashed"}}}, &result); err != nil {
		fmt.Println(err)
	}
}
