// Copyright 2021 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"github.com/simagix/keyhole/mdb"
	"go.mongodb.org/mongo-driver/mongo"
)

// MonitorWiredTigerCache monitor wiredTiger cache
func MonitorWiredTigerCache(version string, client *mongo.Client) {
	wtc := mdb.NewWiredTigerCache(version)
	wtc.Start(client)
}
