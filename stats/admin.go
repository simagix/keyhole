package stats

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// IsMaster - Execute isMaster
func IsMaster(uri string) bson.M {
	session, err := mgo.Dial(uri)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	result := bson.M{}
	if err := session.DB("admin").Run(bson.D{{"isMaster", 1}}, &result); err != nil {
		fmt.Println(err)
	}
	return result
}

// AnalyzeServerStatus -
func AnalyzeServerStatus(filename string) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	var serverStatusData = []ServerStatusData{}
	json.Unmarshal(bytes, &serverStatusData)
	PrintStatsDetails(serverStatusData)
	PrintLatencyDetails(serverStatusData)
	PrintMetricsDetails(serverStatusData)
	PrintWiredTigerDetails(serverStatusData)
}

// PrintStatsDetails -
func PrintStatsDetails(docs []ServerStatusData) {
	stat1 := ServerStatusData{}
	stat2 := ServerStatusData{}
	cnt := 0
	fmt.Println("\n--- Analytic Summary ---")
	fmt.Printf("+-------------------------+-------+-------+------+--------+--------+--------+--------+--------+--------+--------+\n")
	fmt.Printf("| Date/Time               | res   | virt  | fault| Command| Delete | Getmore| Insert | Query  | Update | iops   |\n")
	fmt.Printf("|-------------------------|-------+-------|------|--------|--------|--------|--------|--------|--------|--------|\n")
	for _, doc := range docs {
		bytes, _ := json.Marshal(doc)
		json.Unmarshal(bytes, &stat2)
		if cnt > 0 {
			d := int(stat2.LocalTime.Sub(stat1.LocalTime).Seconds())
			iops := stat2.OpCounters.Command - stat1.OpCounters.Command +
				stat2.OpCounters.Delete - stat1.OpCounters.Delete +
				stat2.OpCounters.Getmore - stat1.OpCounters.Getmore +
				stat2.OpCounters.Insert - stat1.OpCounters.Insert +
				stat2.OpCounters.Query - stat1.OpCounters.Query +
				stat2.OpCounters.Update - stat1.OpCounters.Update
			if d > 0 {
				iops = iops / d
			} else {
				iops = 0
			}
			fmt.Printf("|%-25s|%7d|%7d|%6d|%8d|%8d|%8d|%8d|%8d|%8d|%8d|\n",
				stat2.LocalTime.Format(dateFormat),
				stat2.Mem.Resident,
				stat2.Mem.Virtual,
				stat2.ExtraInfo.PageFaults,
				stat2.OpCounters.Command-stat1.OpCounters.Command,
				stat2.OpCounters.Delete-stat1.OpCounters.Delete,
				stat2.OpCounters.Getmore-stat1.OpCounters.Getmore,
				stat2.OpCounters.Insert-stat1.OpCounters.Insert,
				stat2.OpCounters.Query-stat1.OpCounters.Query,
				stat2.OpCounters.Update-stat1.OpCounters.Update, iops)
		}
		stat1 = stat2
		cnt++
	}
	fmt.Printf("+-------------------------+-------+-------+------+--------+--------+--------+--------+--------+--------+--------+\n")
}

// PrintLatencyDetails -
func PrintLatencyDetails(docs []ServerStatusData) {
	stat1 := ServerStatusData{}
	stat2 := ServerStatusData{}
	cnt := 0
	fmt.Println("\n--- Latencies Summary ---")
	fmt.Printf("+-------------------------+----------+----------+----------+\n")
	fmt.Printf("| Date/Time               | reads    | writes   | commands |\n")
	fmt.Printf("|-------------------------|----------|----------|----------|\n")
	for _, doc := range docs {
		bytes, _ := json.Marshal(doc)
		json.Unmarshal(bytes, &stat2)
		if cnt > 0 {
			r := stat2.OpLatencies.Reads.Ops - stat1.OpLatencies.Reads.Ops
			if r > 0 {
				r = (stat2.OpLatencies.Reads.Latency - stat1.OpLatencies.Reads.Latency) / r
			}
			w := stat2.OpLatencies.Writes.Ops - stat1.OpLatencies.Writes.Ops
			if w > 0 {
				w = (stat2.OpLatencies.Writes.Latency - stat1.OpLatencies.Writes.Latency) / w
			}
			c := stat2.OpLatencies.Commands.Ops - stat1.OpLatencies.Commands.Ops
			if c > 0 {
				c = (stat2.OpLatencies.Commands.Latency - stat1.OpLatencies.Commands.Latency) / c
			}
			fmt.Printf("|%-25s|%10d|%10d|%10d|\n",
				stat2.LocalTime.Format(dateFormat), r, w, c)
		}
		stat1 = stat2
		cnt++
	}
	fmt.Printf("+-------------------------+----------+----------+----------+\n")
}

// PrintMetricsDetails -
func PrintMetricsDetails(docs []ServerStatusData) {
	stat1 := ServerStatusData{}
	stat2 := ServerStatusData{}
	cnt := 0
	fmt.Println("\n--- Metrics ---")
	fmt.Printf("+-------------------------+----------+------------+------------+--------------+----------+----------+----------+----------+\n")
	fmt.Printf("| Date/Time               | Scanned  | ScannedObj |ScanAndOrder|WriteConflicts| Deleted  | Inserted | Returned | Updated  |\n")
	fmt.Printf("|-------------------------|----------|------------|------------|--------------|----------|----------|----------|----------|\n")
	for _, doc := range docs {
		bytes, _ := json.Marshal(doc)
		json.Unmarshal(bytes, &stat2)
		if cnt > 0 {
			fmt.Printf("|%-25s|%10d|%12d|%12d|%14d|%10d|%10d|%10d|%10d|\n",
				stat2.LocalTime.Format(dateFormat),
				stat2.Metrics.QueryExecutor.Scanned-stat1.Metrics.QueryExecutor.Scanned,
				stat2.Metrics.QueryExecutor.ScannedObjects-stat1.Metrics.QueryExecutor.ScannedObjects,
				stat2.Metrics.Operation.ScanAndOrder-stat1.Metrics.Operation.ScanAndOrder,
				stat2.Metrics.Operation.WriteConflicts-stat1.Metrics.Operation.WriteConflicts,
				stat2.Metrics.Document.Deleted-stat1.Metrics.Document.Deleted,
				stat2.Metrics.Document.Inserted-stat1.Metrics.Document.Inserted,
				stat2.Metrics.Document.Returned-stat1.Metrics.Document.Returned,
				stat2.Metrics.Document.Updated-stat1.Metrics.Document.Updated)
		}
		stat1 = stat2
		cnt++
	}
	fmt.Printf("+-------------------------+----------+------------+------------+--------------+----------+----------+----------+----------+\n")
}

// PrintWiredTigerDetails -
func PrintWiredTigerDetails(docs []ServerStatusData) {
	stat1 := ServerStatusData{}
	stat2 := ServerStatusData{}
	cnt := 0
	fmt.Println("\n--- WiredTiger Summary ---")
	fmt.Printf("+-------------------------+--------------------+------------------+------------------------+-------------------+--------------------+-----------------------+\n")
	fmt.Printf("| Date/Time               | MaxBytesConfigured | CurrentlyInCache | UnmodifiedPagesEvicted | TrackedDirtyBytes | PagesReadIntoCache | PagesWrittenFromCache |\n")
	fmt.Printf("|-------------------------|--------------------|------------------|------------------------|-------------------|--------------------|-----------------------|\n")
	for _, doc := range docs {
		bytes, _ := json.Marshal(doc)
		json.Unmarshal(bytes, &stat2)
		if cnt > 0 {
			fmt.Printf("|%-25s|%20d|%18d|%24d|%19d|%20d|%23d|\n",
				stat2.LocalTime.Format(dateFormat),
				stat2.WiredTiger.Cache.MaxBytesConfigured,
				stat2.WiredTiger.Cache.CurrentlyInCache,
				stat2.WiredTiger.Cache.UnmodifiedPagesEvicted,
				stat2.WiredTiger.Cache.TrackedDirtyBytes,
				stat2.WiredTiger.Cache.PagesReadIntoCache-stat1.WiredTiger.Cache.PagesReadIntoCache,
				stat2.WiredTiger.Cache.PagesWrittenFromCache-stat1.WiredTiger.Cache.PagesWrittenFromCache)
		}
		stat1 = stat2
		cnt++
	}
	fmt.Printf("+-------------------------+--------------------+------------------+------------------------+-------------------+--------------------+-----------------------+\n")
}
