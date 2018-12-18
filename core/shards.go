// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/simagix/keyhole/mongo"
)

// Shard information
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
func GetShards(session *mgo.Session, uri string) ([]string, error) {
	var list []string
	result := bson.M{}
	if err := session.DB("admin").Run("listShards", &result); err != nil {
		return list, err
	}

	isSRV := false
	if strings.Index(uri, "mongodb+srv") == 0 {
		isSRV = true
		uri = strings.Replace(uri, "mongodb+srv", "mongodb", 1)
	}
	dialInfo, _ := mongo.ParseURL(uri)
	shards := ShardList{}
	bytes, _ := json.Marshal(result)
	json.Unmarshal(bytes, &shards)

	for s := range shards.Shards {
		host := shards.Shards[s].Host
		s := strings.Split(host, "/")
		if len(s) > 1 {
			dialInfo.ReplicaSetName = s[0]
		}
		ruri := "mongodb://"
		if dialInfo.Username != "" {
			ruri = ruri + dialInfo.Username + ":" + dialInfo.Password + "@"
		}
		if len(s) == 1 {
			ruri = ruri + s[0] + "/?"
		} else {
			ruri = ruri + s[1] + "/?replicaSet=" + dialInfo.ReplicaSetName
		}
		if dialInfo.Source != "" {
			ruri = ruri + "&authSource=" + dialInfo.Source
		}
		if isSRV == true {
			ruri = ruri + "&authSource=admin&ssl=true"
		}
		list = append(list, ruri)
	}
	return list, nil
}

// ShardCollection -
func ShardCollection(session *mgo.Session) error {
	var err error
	// db := session.DB(SimDBName)
	// coll := db.C(CollectionName)
	// coll.DropCollection()
	ssi, err = GetMongoServerInfo(session)
	if err != nil {
		return err
	}
	if ssi.Cluster != SHARDED {
		return nil
	}

	collname := SimDBName + "." + CollectionName
	log.Println("Sharding collection:", collname)
	result := bson.M{}
	if err = session.DB("admin").Run(bson.D{{Name: "enableSharding", Value: SimDBName}}, &result); err != nil {
		return err
	}
	if err = session.DB("admin").Run(bson.D{{Name: "shardCollection", Value: collname}, {Name: "key", Value: bson.M{"_id": "hashed"}}}, &result); err != nil {
		return err
	}
	return nil
}
