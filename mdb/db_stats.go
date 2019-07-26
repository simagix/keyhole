// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

// DBStats contains info from db.stats()
type DBStats struct {
	DB        string `json:"db" bson:"db"`
	DataSize  int    `json:"dataSize" bson:"dataSize"`
	IndexSize int    `json:"indexSize" bson:"indexSize"`
	Objects   int    `json:"objects" bson:"objects"`
}
