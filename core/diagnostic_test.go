package keyhole

import (
	"encoding/json"
	"testing"

	"github.com/globalsign/mgo/bson"
)

func TestReadDiagnosticDir(t *testing.T) {
	var serverInfo interface{}
	var serverStatusList []bson.M
	var err error
	var docs []ServerStatusDoc

	if serverInfo, serverStatusList, err = ReadDiagnosticDir("/tmp/diagnostic.data/"); err != nil {
		t.Fatal(err)
	}
	b, _ := json.MarshalIndent(serverInfo, "", "  ")
	t.Log(string(b))
	for _, ss := range serverStatusList {
		b, _ := json.Marshal(ss)
		doc := ServerStatusDoc{}
		json.Unmarshal(b, &doc)
		docs = append(docs, doc)
	}
	t.Log("serverStatus length", len(serverStatusList), len(docs))
	span := int(docs[(len(docs)-1)].LocalTime.Sub(docs[0].LocalTime).Seconds()) / 20
	if PrintAllStats(docs, span) == "" { // every 10 minutes
		t.Fatal()
	}
}

func TestReadDiagnosticFile(t *testing.T) {
	var serverInfo interface{}
	var serverStatusList []bson.M
	var err error
	var docs []ServerStatusDoc

	if serverInfo, serverStatusList, err = ReadDiagnosticFile("/tmp/metrics.2018-10-12T23-37-51Z-00000"); err != nil {
		t.Fatal(err)
	}
	b, _ := json.MarshalIndent(serverInfo, "", "  ")
	t.Log(string(b))
	for _, ss := range serverStatusList {
		b, _ := json.Marshal(ss)
		doc := ServerStatusDoc{}
		json.Unmarshal(b, &doc)
		docs = append(docs, doc)
	}
	t.Log("serverStatus length", len(serverStatusList), len(docs))
	span := int(docs[(len(docs)-1)].LocalTime.Sub(docs[0].LocalTime).Seconds()) / 20
	if PrintAllStats(docs, span) == "" { // every 10 minutes
		t.Fatal()
	}
}
