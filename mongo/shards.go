// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mongo

import (
	"encoding/json"
	"strings"

	"github.com/globalsign/mgo"
	"github.com/mongodb/mongo-go-driver/bson"
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
func GetShards(session *mgo.Session, uri string) ([]string, error) {
	var err error
	var m bson.M
	var list []string

	if m, err = RunAdminCommand(session, "listShards"); err != nil {
		return list, err
	}

	isSRV := false
	if strings.Index(uri, "mongodb+srv") == 0 {
		isSRV = true
		uri = strings.Replace(uri, "mongodb+srv", "mongodb", 1)
	}
	dialInfo, _ := ParseURL(uri)
	shards := ShardList{}
	bytes, _ := json.Marshal(m)
	json.Unmarshal(bytes, &shards)

	for s := range shards.Shards {
		host := shards.Shards[s].Host
		s := strings.Split(host, "/")
		if len(s) > 1 {
			dialInfo.ReplicaSetName = s[0]
		}
		ruri := "mongodb://"
		if dialInfo.Username != "" {
			ruri += dialInfo.Username + ":" + dialInfo.Password + "@"
		}
		if len(s) == 1 {
			ruri += s[0] + "/?"
		} else {
			ruri += s[1] + "/?replicaSet=" + dialInfo.ReplicaSetName
		}
		if dialInfo.Source != "" {
			ruri += "&authSource=" + dialInfo.Source
		}
		if isSRV == true {
			ruri += "&authSource=admin&ssl=true"
		}
		list = append(list, ruri)
	}
	return list, nil
}
