// Copyright 2019 Kuei-chun Chen. All rights reserved.

package analytics

import (
	"encoding/json"
	"testing"
)

func getServerStatusDocs() []ServerStatusDoc {
	var diag DiagnosticData
	var docs []ServerStatusDoc
	d := NewDiagnosticData()
	diag, _ = d.readDiagnosticFile(DiagnosticDataFilename)

	for _, ss := range diag.ServerStatusList {
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
