// Copyright 2020-present Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"context"
	"strings"
	"sync"

	"github.com/simagix/keyhole/mdb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

var cmutex sync.Mutex

type mongoClientsMap map[string]*mongo.Client

var pool *mongoClientsMap

// GetMongoClientFromPool returns a mongo client by an endpoint
func GetMongoClientFromPool(uri string) (*mongo.Client, error) {
	var err error
	var connstr connstring.ConnString
	cmutex.Lock()
	defer cmutex.Unlock()
	if pool == nil {
		pool = &mongoClientsMap{}
	}
	if (*pool)[uri] == nil {
		if connstr, err = mdb.ParseURI(uri); err != nil {
			return nil, err
		}
		if (*pool)[uri], err = mdb.NewMongoClient(connstr.String()); err != nil {
			(*pool)[uri] = nil
			return (*pool)[uri], err
		}
	} else {
		if err = (*pool)[uri].Ping(context.Background(), nil); err != nil {
			(*pool)[uri] = nil
			return (*pool)[uri], err
		}
	}
	return (*pool)[uri], nil
}

// GetQualifiedDatabases returns a list of qualified database names
func GetQualifiedDatabases(client *mongo.Client) ([]string, error) {
	var err error
	var ctx = context.Background()
	var dbNames = []string{}
	var dbs mdb.ListDatabases
	if err = client.Database("admin").RunCommand(ctx, bson.D{{Key: "listDatabases", Value: 1}}).Decode(&dbs); err != nil {
		return dbNames, err
	}
	for _, db := range dbs.Databases {
		if db.Name == "admin" || db.Name == "config" || db.Name == "local" {
			continue
		}
		dbNames = append(dbNames, db.Name)
	}
	return dbNames, nil
}

// GetQualifiedNamespaces returns a list of qualified namespace names
func GetQualifiedNamespaces(client *mongo.Client) ([]string, error) {
	var err error
	var ctx = context.Background()
	var cursor *mongo.Cursor
	var dbNames []string
	var namespaces = []string{}

	if dbNames, err = GetQualifiedDatabases(client); err != nil {
		return namespaces, err
	}
	for _, dbName := range dbNames {
		if cursor, err = client.Database(dbName).ListCollections(ctx, bson.D{}); err != nil {
			return namespaces, err
		}
		for cursor.Next(ctx) {
			var doc bson.M
			cursor.Decode(&doc)
			name := doc["name"].(string)
			if strings.HasPrefix(name, "system.") && name != "system.js" {
				continue
			} else {
				namespaces = append(namespaces, dbName+"."+name)
			}
		}
		cursor.Close(ctx)
	}
	return namespaces, nil
}

// GetQualifiedNamespacesByDB returns a list of qualified namespace names
func GetQualifiedNamespacesByDB(client *mongo.Client, dbName string) ([]string, error) {
	var err error
	var ctx = context.Background()
	var cursor *mongo.Cursor
	var namespaces = []string{}

	if cursor, err = client.Database(dbName).ListCollections(ctx, bson.D{}); err != nil {
		return namespaces, err
	}
	for cursor.Next(ctx) {
		var doc bson.M
		cursor.Decode(&doc)
		name := doc["name"].(string)
		if strings.HasPrefix(name, "system.") && name != "system.js" {
			continue
		} else {
			namespaces = append(namespaces, dbName+"."+name)
		}
	}
	cursor.Close(ctx)
	return namespaces, nil
}
