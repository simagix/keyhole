package mongo

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

var mb = 1024.0 * 1024
var mongoStats = make(map[string]interface{})
var sleepTime = 10 * time.Second

// WiredTigerData -
type WiredTigerData struct {
	Perf interface{}
}

// Document -
type Document struct {
	Deleted  int `json:"deleted" bson:"deleted"`
	Inserted int `json:"inserted" bson:"inserted"`
	Returned int `json:"returned" bson:"returned"`
	Updated  int `json:"updated" bson:"updated"`
}

// Metrics -
type Metrics struct {
	Document Document `json:"document" bson:"document"`
}

// Mem -
type Mem struct {
	Resident int `json:"resident" bson:"resident"`
	Virtual  int `json:"virtual" bson:"virtual"`
}

// ExtraInfo -
type ExtraInfo struct {
	PageFaults int `json:"page_faults" bson:"page_faults"`
}

// ServerStatusData -
type ServerStatusData struct {
	Mem        Mem       `json:"Mem" bson:"Mem"`
	ExtraInfo  ExtraInfo `json:"extra_info" bson:"extra_info"`
	Metrics    Metrics   `json:"metrics" bson:"metrics"`
	WiredTiger WiredTigerData
}

// CollectServerStatus - Collect serverStatus every 10 minutes
func CollectServerStatus(uri string) {
	stat := ServerStatusData{}
	var iop int
	var piop int
	for {
		bytes, _ := json.Marshal(ServerStatus(uri))
		json.Unmarshal(bytes, &stat)
		key := time.Now().Format("2006-01-02T15:04:05-07:00")
		mongoStats[key] = stat
		fmt.Printf("%s resident: %d, virtual: %d, page faults: %d\n",
			key, stat.Mem.Resident, stat.Mem.Virtual, stat.ExtraInfo.PageFaults)
		iop = stat.Metrics.Document.Inserted + stat.Metrics.Document.Returned +
			stat.Metrics.Document.Updated + stat.Metrics.Document.Deleted
		iops := (iop - piop) / 60
		if piop == 0 {
			iops = 0
		}
		fmt.Printf("%s metrics: c: %d, r: %d, u: %d, d: %d, iops: %d\n",
			key, stat.Metrics.Document.Inserted, stat.Metrics.Document.Returned,
			stat.Metrics.Document.Updated, stat.Metrics.Document.Deleted, iops)
		piop = iop
		time.Sleep(1 * time.Minute)
	}
}

// PrintDBStats - Print dbStats every 10 seconds
func PrintDBStats(uri string) {
	var raw map[string]interface{}
	var pds float64
	var ds float64
	ptime := time.Now()
	now := ptime

	for {
		stat := DBStats(uri)
		bytes, _ := json.Marshal(stat)
		json.Unmarshal(bytes, &raw)
		ds = raw["dataSize"].(float64)
		sec := now.Sub(ptime).Seconds()
		delta := (ds - pds) / mb / sec
		if sec > 0 && delta > .01 {
			fmt.Printf("%s %8.1f -> %8.1f, rate %8.1f MB/second\n",
				now.Format("2006-01-02T15:04:05-07:00"), pds/mb, ds/mb, delta)
		}
		pds = ds
		ptime = now
		time.Sleep(sleepTime)
		now = time.Now()
	}
}

// PrintServerStatus - Print serverStatus summary for the duration
func PrintServerStatus(uri string) {
	stat := ServerStatusData{}
	bytes, _ := json.Marshal(ServerStatus(uri))
	json.Unmarshal(bytes, &stat)
	mongoStats[time.Now().Format("2006-01-02T15:04:05-07:00")] = stat

	var keys []string
	for k := range mongoStats {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	key := keys[0]
	bytes, _ = json.MarshalIndent(mongoStats[key], "", "  ")
	fmt.Println("\nKey:", key, "Value:", string(bytes))
	key = keys[len(keys)-1]
	bytes, _ = json.MarshalIndent(mongoStats[key], "", "  ")
	fmt.Println("Key:", key, "Value:", string(bytes))
}
