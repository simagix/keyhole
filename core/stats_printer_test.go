// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"encoding/json"
	"testing"
)

const filename = "/tmp/metrics.2018-10-18T12-06-57Z-00000"

func getServerStatusDocs() []ServerStatusDoc {
	var docs []ServerStatusDoc
	d := NewDiagnosticData(false)
	d.readDiagnosticFile(filename)

	for _, ss := range d.ServerStatusList {
		b, _ := json.Marshal(ss)
		doc := ServerStatusDoc{}
		json.Unmarshal(b, &doc)
		docs = append(docs, doc)
	}
	return docs
}

func TestPrintWiredTigerConcurrentTransactionsDetails(t *testing.T) {
	docs := getServerStatusDocs()
	printWiredTigerConcurrentTransactionsDetails(docs, 600) // every 10 minutes
	span := int(docs[(len(docs)-1)].LocalTime.Sub(docs[0].LocalTime).Seconds()) / 20
	t.Log(printWiredTigerConcurrentTransactionsDetails(docs, span))
}

func TestPrintWiredTigerCacheDetails(t *testing.T) {
	docs := getServerStatusDocs()
	printWiredTigerCacheDetails(docs, 600) // every 10 minutes
	span := int(docs[(len(docs)-1)].LocalTime.Sub(docs[0].LocalTime).Seconds()) / 20
	t.Log(printWiredTigerCacheDetails(docs, span))
}

func TestPrintGlobalLockDetails(t *testing.T) {
	docs := getServerStatusDocs()
	printGlobalLockDetails(docs, 600) // every 10 minutes
	span := int(docs[(len(docs)-1)].LocalTime.Sub(docs[0].LocalTime).Seconds()) / 20
	t.Log(printGlobalLockDetails(docs, span))
}

func TestPrintMetricsDetails(t *testing.T) {
	docs := getServerStatusDocs()
	printMetricsDetails(docs, 600) // every 10 minutes
	span := int(docs[(len(docs)-1)].LocalTime.Sub(docs[0].LocalTime).Seconds()) / 20
	t.Log(printMetricsDetails(docs, span))
}

func TestPrintLatencyDetails(t *testing.T) {
	docs := getServerStatusDocs()
	printLatencyDetails(docs, 600) // every 10 minutes
	span := int(docs[(len(docs)-1)].LocalTime.Sub(docs[0].LocalTime).Seconds()) / 20
	t.Log(printLatencyDetails(docs, span))
}

func TestPrintStatsDetails(t *testing.T) {
	docs := getServerStatusDocs()
	printStatsDetails(docs, 600) // every 10 minutes
	span := int(docs[(len(docs)-1)].LocalTime.Sub(docs[0].LocalTime).Seconds()) / 20
	t.Log(printStatsDetails(docs, span))
}
