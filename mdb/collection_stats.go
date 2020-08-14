// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"go.mongodb.org/mongo-driver/bson"
)

// CollectionStats stores struct
type CollectionStats struct {
	Collection string       `bson:"collection"`
	Chunks     []ChunkStats `bson:"chunks"`
	Document   bson.M       `bson:"document"`
	Indexes    []IndexStats `bson:"indexes"`
	NS         string       `bson:"NS"`
	Stats      struct {
		Count          int64   `bson:"count,truncate"`
		IndexDetails   bson.M  `bson:"indexDetails"`
		AvgObjSize     float64 `bson:"avgObjSize"`
		WiredTiger     bson.M  `bson:"wiredTiger"`
		IndexSizes     bson.M  `bson:"indexSizes"`
		Capped         bool    `bson:"capped"`
		TotalIndexSize int64   `bson:"totalIndexSize,truncate"`
		MaxSize        int64   `bson:"maxSize,truncate"`
		Nindexes       int64   `bson:"nindexes,truncate"`
		Shards         bson.M  `bson:"shards"`
		NS             string  `bson:"ns"`
		Nchunks        int64   `bson:"nchunks,truncate"`
		Sharded        bool    `bson:"sharded"`
		Size           int64   `bson:"size,truncate"`
		StorageSize    int64   `bson:"storageSize,truncate"`
	} `bson:"stats"`
}
