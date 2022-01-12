// Copyright 2020-present Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/simagix/gox"

	"github.com/simagix/keyhole/mdb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

// Logv2Network stores logv2 network info
type Logv2Network struct {
	Attributes struct {
		ConnectionCount int    `json:"connectionCount" bson:"connectionCount"`
		ConnectionID    int    `json:"connectionId" bson:"connectionId"`
		Remote          string `json:"remote" bson:"remote"`
	} `json:"attr" bson:"attr"`
	Component string            `json:"c" bson:"c"`
	Context   string            `json:"ctx" bson:"ctx"`
	ID        int               `json:"id" bson:"id"`
	Message   string            `json:"msg" bson:"msg"`
	Timestamp map[string]string `json:"t" bson:"t"`
}

// PrintConnections print all connection info
func PrintConnections(cfg *Config) error {
	var err error
	if cfg.Filename != "" {
		if err = PrintConnectionsFromFile(cfg.Filename); err != nil {
			return err
		}
	}
	if cfg.URI != "" {
		if err = PrintConnectionsFromURI(cfg.URI); err != nil {
			return err
		}
	}
	return err
}

// PrintConnectionsFromFile print all connection info from a log file
func PrintConnectionsFromFile(filename string) error {
	var err error
	var reader *bufio.Reader
	if reader, err = gox.NewFileReader(filename); err != nil {
		return err
	}
	accepted := 0
	ended := 0
	connMap := map[string][2]int{}
	for {
		var data []byte
		if data, _, err = reader.ReadLine(); err != nil { // 0x0A separator = newline
			break
		}
		str := string(data)
		if !strings.Contains(str, `"c":"NETWORK"`) {
			continue
		}
		var doc Logv2Network
		if err = json.Unmarshal(data, &doc); err != nil || doc.Component != "NETWORK" {
			continue
		}
		strs := strings.Split(doc.Attributes.Remote, ":")
		if len(connMap[strs[0]]) == 0 {
			connMap[strs[0]] = [2]int{0, 0}
		}
		nums := connMap[strs[0]]
		if doc.Message == "Connection accepted" {
			accepted++
			fmt.Printf(" - conn ID: %-10d desc: conn%-10d client: %-40s\n", doc.Attributes.ConnectionID, doc.Attributes.ConnectionID, doc.Attributes.Remote)
			nums[0]++
			connMap[strs[0]] = nums
		} else if doc.Message == "Connection ended" {
			ended++
			nums[1]++
			connMap[strs[0]] = nums
		}
	}
	fmt.Println()
	fmt.Println("Summary:")
	fmt.Println(" - connections accepted", accepted)
	fmt.Println(" - connections ended", ended)
	for k, v := range connMap {
		fmt.Printf(" - %v had %v accepted, %v ended\n", k, v[0], v[1])
	}
	return nil
}

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
