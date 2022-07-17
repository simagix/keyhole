// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"log"
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

func TestOutputBSON(t *testing.T) {
	uri := "mongodb://admin:secret@localhost:30309/"
	if os.Getenv("DATABASE_URL") != "" {
		uri = os.Getenv("DATABASE_URL")
	}
	var connString connstring.ConnString
	connString, err := ParseURI(uri)
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewMongoClient(uri)
	if err != nil {
		t.Fatal(err)
	}
	stats := NewClusterStats("simagix/keyhole test")
	if err = stats.GetClusterStats(client, connString); err != nil {
		log.Fatalf("error GetClusterStats: %v", err)
	}

	// test bson.Marshal() overflow
	// for index, db := range *stats.Databases {
	// 	for i := 0; i < 12; i++ {
	// 		colls := db.Collections
	// 		(*stats.Databases)[index].Collections = append(db.Collections, colls...)
	// 	}
	// }

	if _, _, err = stats.OutputBSON(); err != nil {
		log.Fatalf("failed to output bson file\n%v", err)
	}
}
