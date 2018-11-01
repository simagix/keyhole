package charts

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	keyhole "github.com/simagix/keyhole/core"
)

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path[1:] == "memory" || r.URL.Path[1:] == "" {
		str := strings.Replace(IndexHTML, "__TITLE__", "Memory (GB)", -1)
		str = strings.Replace(str, "__API__", "v1/memory/tsv", -1)
		fmt.Fprintf(w, str)
	} else if r.URL.Path[1:] == "v1/memory/tsv" {
		fmt.Fprintf(w, strings.Join(GetMemoryTSV()[:], "\n"))

	} else if r.URL.Path[1:] == "page_faults" {
		str := strings.Replace(IndexHTML, "__TITLE__", "Page Faults", -1)
		str = strings.Replace(str, "__API__", "v1/page_faults/tsv", -1)
		fmt.Fprintf(w, str)
	} else if r.URL.Path[1:] == "v1/page_faults/tsv" {
		fmt.Fprintf(w, strings.Join(GetPageFaultsTSV()[:], "\n"))

	} else if r.URL.Path[1:] == "metrics" {
		str := strings.Replace(IndexHTML, "__TITLE__", "Metrics", -1)
		str = strings.Replace(str, "__API__", "v1/metrics/tsv", -1)
		fmt.Fprintf(w, str)
	} else if r.URL.Path[1:] == "v1/metrics/tsv" {
		fmt.Fprintf(w, strings.Join(GetMetricsTSV()[:], "\n"))

	} else if r.URL.Path[1:] == "wiredtiger_cache" {
		str := strings.Replace(IndexHTML, "__TITLE__", "WiredTiger Cache (GB)", -1)
		str = strings.Replace(str, "__API__", "v1/wiredtiger_cache/tsv", -1)
		fmt.Fprintf(w, str)
	} else if r.URL.Path[1:] == "v1/wiredtiger_cache/tsv" {
		fmt.Fprintf(w, strings.Join(GetWiredTigerCacheTSV()[:], "\n"))

	} else if r.URL.Path[1:] == "ops" {
		str := strings.Replace(IndexHTML, "__TITLE__", "Ops Counters", -1)
		str = strings.Replace(str, "__API__", "v1/ops/tsv", -1)
		fmt.Fprintf(w, str)
	} else if r.URL.Path[1:] == "v1/ops/tsv" {
		fmt.Fprintf(w, strings.Join(GetOpCountersTSV()[:], "\n"))

	} else if r.URL.Path[1:] == "wiredtiger_tickets" {
		str := strings.Replace(IndexHTML, "__TITLE__", "WiredTiger Tickets", -1)
		str = strings.Replace(str, "__API__", "v1/wiredtiger_tickets/tsv", -1)
		fmt.Fprintf(w, str)
	} else if r.URL.Path[1:] == "v1/wiredtiger_tickets/tsv" {
		fmt.Fprintf(w, strings.Join(GetWiredTigerTicketsTSV()[:], "\n"))

	} else if r.URL.Path[1:] == "wiredtiger_paging" {
		str := strings.Replace(IndexHTML, "__TITLE__", "WiredTiger Paging (pages per minute)", -1)
		str = strings.Replace(str, "__API__", "v1/wiredtiger_paging/tsv", -1)
		fmt.Fprintf(w, str)
	} else if r.URL.Path[1:] == "v1/wiredtiger_paging/tsv" {
		fmt.Fprintf(w, strings.Join(GetWiredTigerPagingTSV()[:], "\n"))

	} else if r.URL.Path[1:] == "latencies" {
		str := strings.Replace(IndexHTML, "__TITLE__", "Latencies (milliseconds)", -1)
		str = strings.Replace(str, "__API__", "v1/latencies/tsv", -1)
		fmt.Fprintf(w, str)
	} else if r.URL.Path[1:] == "v1/latencies/tsv" {
		fmt.Fprintf(w, strings.Join(GetLatenciesTSV()[:], "\n"))

	} else if r.URL.Path[1:] == "connections" {
		str := strings.Replace(IndexHTML, "__TITLE__", "Connections", -1)
		str = strings.Replace(str, "__API__", "v1/connections/tsv", -1)
		fmt.Fprintf(w, str)
	} else if r.URL.Path[1:] == "v1/connections/tsv" {
		fmt.Fprintf(w, strings.Join(GetConnectionsTSV()[:], "\n"))

	} else {
		fmt.Fprintf(w, "Keyhole Performance Charts!  Unknow API!")
	}
}

// HTTPServer listens to port 5408
func HTTPServer(port int) {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}

// GetMemoryTSV -
func GetMemoryTSV() []string {
	var docs []string
	var r, v float64
	docs = append(docs, "date\tResident\tVirtual")
	for _, value := range keyhole.ChartsDocs {
		stat := keyhole.ServerStatusDoc{}
		for _, doc := range value {
			buf, _ := json.Marshal(doc)
			json.Unmarshal(buf, &stat)
			r = float64(stat.Mem.Resident) / 1024
			v = float64(stat.Mem.Virtual) / 1024
			docs = append(docs, stat.LocalTime.Format("2006-01-02T15:04:05Z")+
				"\t"+strconv.FormatFloat(r, 'E', -1, 64)+
				"\t"+strconv.FormatFloat(v, 'E', -1, 64))
		}
		break
	}

	return docs
}

// GetPageFaultsTSV -
func GetPageFaultsTSV() []string {
	var docs []string
	pstat := keyhole.ServerStatusDoc{}
	docs = append(docs, "date\tPage Faults")
	for _, value := range keyhole.ChartsDocs {
		stat := keyhole.ServerStatusDoc{}
		for i, doc := range value {
			buf, _ := json.Marshal(doc)
			json.Unmarshal(buf, &stat)
			if i > 0 && stat.Uptime > pstat.Uptime {
				n := stat.ExtraInfo.PageFaults - pstat.ExtraInfo.PageFaults
				docs = append(docs, stat.LocalTime.Format("2006-01-02T15:04:05Z")+"\t"+strconv.Itoa(n))
			}
			pstat = stat
		}
		break
	}

	return docs
}

// GetMetricsTSV -
func GetMetricsTSV() []string {
	var docs []string
	var s1, s2, s3 int
	var d, i, r, u int
	var stat, pstat keyhole.ServerStatusDoc

	docs = append(docs, "date\tScanned Keys\tScanned Objects\tScan And Order\tDeleted\tInserted\tReturned\tUpdated")
	for _, value := range keyhole.ChartsDocs {
		for _, doc := range value {
			buf, _ := json.Marshal(doc)
			json.Unmarshal(buf, &stat)
			if pstat.Host != "" && stat.Uptime > pstat.Uptime {
				s1 = stat.Metrics.QueryExecutor.Scanned - pstat.Metrics.QueryExecutor.Scanned
				s2 = stat.Metrics.QueryExecutor.ScannedObjects - pstat.Metrics.QueryExecutor.ScannedObjects
				s3 = stat.Metrics.Operation.ScanAndOrder - pstat.Metrics.Operation.ScanAndOrder
				d = stat.Metrics.Document.Deleted - pstat.Metrics.Document.Deleted
				i = stat.Metrics.Document.Inserted - pstat.Metrics.Document.Inserted
				r = stat.Metrics.Document.Returned - pstat.Metrics.Document.Returned
				u = stat.Metrics.Document.Updated - pstat.Metrics.Document.Updated
				docs = append(docs, stat.LocalTime.Format("2006-01-02T15:04:05Z")+
					"\t"+strconv.Itoa(s1)+"\t"+strconv.Itoa(s2)+"\t"+strconv.Itoa(s3)+
					"\t"+strconv.Itoa(d)+"\t"+strconv.Itoa(i)+"\t"+strconv.Itoa(r)+"\t"+strconv.Itoa(u))
			}
			pstat = stat
		}
	}

	return docs
}

// GetWiredTigerCacheTSV -
func GetWiredTigerCacheTSV() []string {
	var docs []string
	var m, c, t float64

	docs = append(docs, "date\tMax Bytes\tIn Cache\tDirty Bytes")
	for _, value := range keyhole.ChartsDocs {
		stat := keyhole.ServerStatusDoc{}
		for _, doc := range value {
			buf, _ := json.Marshal(doc)
			json.Unmarshal(buf, &stat)
			m = float64(stat.WiredTiger.Cache.MaxBytesConfigured) / (1024 * 1024 * 1024)
			c = float64(stat.WiredTiger.Cache.CurrentlyInCache) / (1024 * 1024 * 1024)
			t = float64(stat.WiredTiger.Cache.TrackedDirtyBytes) / (1024 * 1024 * 1024)
			docs = append(docs, stat.LocalTime.Format("2006-01-02T15:04:05Z")+
				"\t"+strconv.FormatFloat(m, 'E', -1, 64)+
				"\t"+strconv.FormatFloat(c, 'E', -1, 64)+
				"\t"+strconv.FormatFloat(t, 'E', -1, 64))
		}
		break
	}

	return docs
}

// GetWiredTigerPagingTSV -
func GetWiredTigerPagingTSV() []string {
	var docs []string
	var stat, pstat keyhole.ServerStatusDoc
	var m, u, r, w float64

	docs = append(docs, "date\tModified Evicted\tUnmodified Evicted\tRead In Cache\tWritten From Cache")
	cnt := 0
	for _, value := range keyhole.ChartsDocs {
		for _, doc := range value {
			buf, _ := json.Marshal(doc)
			json.Unmarshal(buf, &stat)
			if cnt > 0 && stat.Uptime > pstat.Uptime {
				minutes := stat.LocalTime.Sub(pstat.LocalTime).Minutes()
				m = float64(stat.WiredTiger.Cache.ModifiedPagesEvicted-pstat.WiredTiger.Cache.ModifiedPagesEvicted) / minutes
				u = float64(stat.WiredTiger.Cache.UnmodifiedPagesEvicted-pstat.WiredTiger.Cache.UnmodifiedPagesEvicted) / minutes
				r = float64(stat.WiredTiger.Cache.PagesReadIntoCache-pstat.WiredTiger.Cache.PagesReadIntoCache) / minutes
				w = float64(stat.WiredTiger.Cache.PagesWrittenFromCache-pstat.WiredTiger.Cache.PagesWrittenFromCache) / minutes
				docs = append(docs, stat.LocalTime.Format("2006-01-02T15:04:05Z")+
					"\t"+strconv.FormatFloat(m, 'E', -1, 64)+
					"\t"+strconv.FormatFloat(u, 'E', -1, 64)+
					"\t"+strconv.FormatFloat(r, 'E', -1, 64)+
					"\t"+strconv.FormatFloat(w, 'E', -1, 64))
			}
			pstat = stat
			cnt++
		}
	}

	return docs
}

// GetWiredTigerTicketsTSV -
func GetWiredTigerTicketsTSV() []string {
	var docs []string
	docs = append(docs, "date\tRead Ticket Available\tWrite Ticket Available")
	for _, value := range keyhole.ChartsDocs {
		stat := keyhole.ServerStatusDoc{}
		for _, doc := range value {
			buf, _ := json.Marshal(doc)
			json.Unmarshal(buf, &stat)
			docs = append(docs, stat.LocalTime.Format("2006-01-02T15:04:05Z")+
				"\t"+strconv.Itoa(stat.WiredTiger.ConcurrentTransactions.Read.Available)+
				"\t"+strconv.Itoa(stat.WiredTiger.ConcurrentTransactions.Write.Available))
		}
		break
	}

	return docs
}

// GetOpCountersTSV -
func GetOpCountersTSV() []string {
	var docs []string
	pstat := keyhole.ServerStatusDoc{}
	docs = append(docs, "date\tQuery\tInsert\tUpdate\tDelete\tGet More\tCommand")
	for _, value := range keyhole.ChartsDocs {
		stat := keyhole.ServerStatusDoc{}
		for i, doc := range value {
			buf, _ := json.Marshal(doc)
			json.Unmarshal(buf, &stat)
			if i > 0 && stat.Uptime > pstat.Uptime {
				qry := stat.OpCounters.Query - pstat.OpCounters.Query
				ins := stat.OpCounters.Insert - pstat.OpCounters.Insert
				upt := stat.OpCounters.Update - pstat.OpCounters.Update
				del := stat.OpCounters.Delete - pstat.OpCounters.Delete
				gm := stat.OpCounters.Getmore - pstat.OpCounters.Getmore
				cmd := stat.OpCounters.Command - pstat.OpCounters.Command

				docs = append(docs, stat.LocalTime.Format("2006-01-02T15:04:05Z")+"\t"+
					strconv.Itoa(qry)+"\t"+strconv.Itoa(ins)+"\t"+strconv.Itoa(upt)+"\t"+
					strconv.Itoa(del)+"\t"+strconv.Itoa(gm)+"\t"+strconv.Itoa(cmd))
			}
			pstat = stat
		}
		break
	}

	return docs
}

// GetLatenciesTSV -
func GetLatenciesTSV() []string {
	var docs []string
	var r, w, c float64
	// var pstat keyhole.ServerStatusDoc

	docs = append(docs, "date\tReads (ms)\tWrites (ms)\tCommands (ms)")
	for _, value := range keyhole.ChartsDocs {
		stat := keyhole.ServerStatusDoc{}
		for _, doc := range value {
			buf, _ := json.Marshal(doc)
			json.Unmarshal(buf, &stat)
			r = 0
			if stat.OpLatencies.Reads.Ops > 0 {
				r = float64(stat.OpLatencies.Reads.Latency) / float64(stat.OpLatencies.Reads.Ops) / 1000
			}
			w = 0
			if stat.OpLatencies.Writes.Ops > 0 {
				w = float64(stat.OpLatencies.Writes.Latency) / float64(stat.OpLatencies.Writes.Ops) / 1000
			}
			c = 0
			if stat.OpLatencies.Commands.Ops > 0 {
				c = float64(stat.OpLatencies.Commands.Latency) / float64(stat.OpLatencies.Commands.Ops) / 1000
			}
			docs = append(docs, stat.LocalTime.Format("2006-01-02T15:04:05Z")+
				"\t"+strconv.FormatFloat(r, 'E', -1, 64)+
				"\t"+strconv.FormatFloat(w, 'E', -1, 64)+
				"\t"+strconv.FormatFloat(c, 'E', -1, 64))
			// pstat = stat
		}
		break
	}
	return docs
}

// GetConnectionsTSV -
func GetConnectionsTSV() []string {
	var docs []string
	var stat, pstat keyhole.ServerStatusDoc

	docs = append(docs, "date\tCurrent\tAvailable\tCreated per minute")
	for _, value := range keyhole.ChartsDocs {
		for _, doc := range value {
			buf, _ := json.Marshal(doc)
			json.Unmarshal(buf, &stat)
			if pstat.Host != "" && stat.Uptime > pstat.Uptime {
				minutes := stat.LocalTime.Sub(pstat.LocalTime).Minutes()
				churn := float64(stat.Connections.TotalCreated-pstat.Connections.TotalCreated) / minutes
				docs = append(docs, stat.LocalTime.Format("2006-01-02T15:04:05Z")+
					"\t"+strconv.Itoa(stat.Connections.Current)+
					"\t"+strconv.Itoa(stat.Connections.Available)+
					"\t"+strconv.Itoa(int(churn)))
			}
			pstat = stat
		}
		break
	}

	return docs
}
