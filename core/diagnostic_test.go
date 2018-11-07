package keyhole

import (
	"encoding/json"
	"testing"
)

func TestReadDiagnosticDir(t *testing.T) {
	var serverInfo interface{}
	var serverStatusList []ServerStatusDoc
	var replSetStatusList []ReplSetStatusDoc
	var err error

	if serverInfo, serverStatusList, replSetStatusList, err = ReadDiagnosticDir("/tmp/diagnostic.data/"); err != nil {
		t.Fatal(err)
	}
	b, _ := json.MarshalIndent(serverInfo, "", "  ")
	t.Log(string(b))
	t.Log("serverStatus length", len(serverStatusList))
	t.Log("replSetStatus length", len(replSetStatusList))
	span := int(serverStatusList[(len(serverStatusList)-1)].LocalTime.Sub(serverStatusList[0].LocalTime).Seconds()) / 20
	if PrintAllStats(serverStatusList, span) == "" { // every 10 minutes
		t.Fatal()
	}
}

func TestReadDiagnosticFile(t *testing.T) {
	var serverInfo interface{}
	var serverStatusList []ServerStatusDoc
	var replSetStatusList []ReplSetStatusDoc
	var err error

	if serverInfo, serverStatusList, replSetStatusList, err = ReadDiagnosticFile("/tmp/metrics.2018-10-12T23-37-51Z-00000"); err != nil {
		t.Fatal(err)
	}
	b, _ := json.MarshalIndent(serverInfo, "", "  ")
	t.Log(string(b))
	t.Log("serverStatus length", len(serverStatusList))
	t.Log("replSetStatus length", len(replSetStatusList))
	span := int(serverStatusList[(len(serverStatusList)-1)].LocalTime.Sub(serverStatusList[0].LocalTime).Seconds()) / 20
	if PrintAllStats(serverStatusList, span) == "" { // every 10 minutes
		t.Fatal()
	}
}
