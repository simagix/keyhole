// Copyright 2020-present Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/simagix/keyhole/mdb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

// PrintConnectionsFromURI print all connection info from all mongod
func PrintConnectionsFromURI(uri string) error {
	var err error
	var client *mongo.Client
	var ctx = context.Background()
	pipeline := `[{
		"$currentOp": {
			"allUsers": true,
			"idleSessions": true
		}
	}, {
		"$match": {
			"connectionId": {
				"$exists": true
			}
		}
	}, {
		"$group": {
			"_id": {
				"id": "$connectionId",
				"desc": "$desc",
				"client": "$client"
			}
		}
	}]`

	// connection string is required from here forward
	var connString connstring.ConnString
	if connString, err = mdb.ParseURI(uri); err != nil {
		return err
	}
	uri = connString.String() // password can be injected if missing
	if client, err = mdb.NewMongoClient(uri); err != nil {
		return err
	}
	stats := mdb.NewClusterStats("keyhole")
	if err = stats.GetClusterStatsSummary(client); err != nil {
		fmt.Println(err)
		return err
	}

	var uris []string
	uriMap := map[string][]string{}
	if stats.Cluster == mdb.Replica {
		setName := stats.ServerStatus.Repl.SetName
		s := fmt.Sprintf(`%v/%v`, setName, strings.Join(stats.ServerStatus.Repl.Hosts, ","))
		oneShard := []mdb.Shard{{ID: setName, State: 1, Host: s}}
		if uris, err = mdb.GetAllServerURIs(oneShard, connString); err != nil {
			return err
		}
		uriMap[setName] = uris
	} else if stats.Cluster == mdb.Sharded {
		for _, shard := range stats.Shards {
			if uris, err = mdb.GetAllServerURIs([]mdb.Shard{shard}, connString); err != nil {
				return err
			}
			uriMap[shard.ID] = uris
		}
	}
	for setName, uris := range uriMap {
		for _, conn := range uris {
			if client, err = mdb.NewMongoClient(conn); err != nil {
				return err
			}
			var cursor *mongo.Cursor
			opts := options.Aggregate().SetAllowDiskUse(true)
			if cursor, err = client.Database("admin").Aggregate(ctx, mdb.MongoPipeline(pipeline), opts); err != nil {
				log.Println(err)
				continue
			}
			if connString, err = mdb.ParseURI(conn); err != nil {
				return err
			}

			fmt.Printf("%v/%v\n", setName, connString.Hosts[0])
			for cursor.Next(ctx) {
				var doc bson.M
				if err = cursor.Decode(&doc); err != nil {
					log.Println(err)
					continue
				}
				conn := doc["_id"].(bson.M)
				fmt.Printf(" - conn ID: %-10d desc: %-20s client: %-40s\n", conn["id"], conn["desc"], conn["client"])
			}
			cursor.Close(ctx)
		}
	}
	return err
}
