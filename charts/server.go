package charts

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/simagix/keyhole/stats"
)

func handler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path[1:] == "memory" {
		str := strings.Replace(IndexHTML, "__TITLE__", "Memory Stats Charts", -1)
		fmt.Fprintf(w, strings.Replace(str, "__MODULE__", "memory", -1))
	} else if r.URL.Path[1:] == "memory/index.js" {
		fmt.Fprintf(w, strings.Replace(D3JS, "__API__", "v1/memory/tsv", -1))
	} else if r.URL.Path[1:] == "v1/memory/tsv" {
		fmt.Fprintf(w, strings.Join(GetMemoryTSV()[:], "\n"))

	} else if r.URL.Path[1:] == "page_faults" {
		str := strings.Replace(IndexHTML, "__TITLE__", "Page Faults Charts", -1)
		fmt.Fprintf(w, strings.Replace(str, "__MODULE__", "page_faults", -1))
	} else if r.URL.Path[1:] == "page_faults/index.js" {
		fmt.Fprintf(w, strings.Replace(D3JS, "__API__", "v1/page_faults/tsv", -1))
	} else if r.URL.Path[1:] == "v1/page_faults/tsv" {
		fmt.Fprintf(w, strings.Join(GetPageFaultsTSV()[:], "\n"))

	} else if r.URL.Path[1:] == "wiredtiger_cache" {
		str := strings.Replace(IndexHTML, "__TITLE__", "WiredTiger Cache Charts", -1)
		fmt.Fprintf(w, strings.Replace(str, "__MODULE__", "wiredtiger_cache", -1))
	} else if r.URL.Path[1:] == "wiredtiger_cache/index.js" {
		fmt.Fprintf(w, strings.Replace(D3JS, "__API__", "v1/wiredtiger_cache/tsv", -1))
	} else if r.URL.Path[1:] == "v1/wiredtiger_cache/tsv" {
		fmt.Fprintf(w, strings.Join(GetWiredTigerCacheTSV()[:], "\n"))

	} else if r.URL.Path[1:] == "wiredtiger_tickets" {
		str := strings.Replace(IndexHTML, "__TITLE__", "WiredTiger Concurrent Transactions Charts", -1)
		fmt.Fprintf(w, strings.Replace(str, "__MODULE__", "wiredtiger_tickets", -1))
	} else if r.URL.Path[1:] == "wiredtiger_tickets/index.js" {
		fmt.Fprintf(w, strings.Replace(D3JS, "__API__", "v1/wiredtiger_tickets/tsv", -1))
	} else if r.URL.Path[1:] == "v1/wiredtiger_tickets/tsv" {
		fmt.Fprintf(w, strings.Join(GetWiredTigerTicketsTSV()[:], "\n"))

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
	docs = append(docs, "date\tResident")
	for _, value := range stats.ChartsDocs {
		stat := stats.ServerStatusDoc{}
		for _, doc := range value {
			buf, _ := json.Marshal(doc)
			json.Unmarshal(buf, &stat)
			docs = append(docs, stat.LocalTime.Format("2006-01-02T15:04:05Z")+"\t"+strconv.Itoa(stat.Mem.Resident))
		}
		break
	}

	return docs
}

// GetPageFaultsTSV -
func GetPageFaultsTSV() []string {
	var docs []string
	pstat := stats.ServerStatusDoc{}
	docs = append(docs, "date\tPage Faults")
	for _, value := range stats.ChartsDocs {
		stat := stats.ServerStatusDoc{}
		for i, doc := range value {
			buf, _ := json.Marshal(doc)
			json.Unmarshal(buf, &stat)
			n := stat.ExtraInfo.PageFaults - pstat.ExtraInfo.PageFaults
			if i > 0 {
				docs = append(docs, stat.LocalTime.Format("2006-01-02T15:04:05Z")+"\t"+strconv.Itoa(n))
			}
			pstat = stat
		}
		break
	}

	return docs
}

// GetWiredTigerCacheTSV -
func GetWiredTigerCacheTSV() []string {
	var docs []string
	docs = append(docs, "date\tMax Bytes(MB)\tIn Cache(MB)\tDirty Bytes(MB)")
	for _, value := range stats.ChartsDocs {
		stat := stats.ServerStatusDoc{}
		for _, doc := range value {
			buf, _ := json.Marshal(doc)
			json.Unmarshal(buf, &stat)
			docs = append(docs, stat.LocalTime.Format("2006-01-02T15:04:05Z")+"\t"+strconv.Itoa(stat.WiredTiger.Cache.MaxBytesConfigured/(1024*1024))+"\t"+strconv.Itoa(stat.WiredTiger.Cache.CurrentlyInCache/(1024*1024))+"\t"+strconv.Itoa(stat.WiredTiger.Cache.TrackedDirtyBytes/(1024*1024)))
		}
		break
	}

	return docs
}

// GetWiredTigerTicketsTSV -
func GetWiredTigerTicketsTSV() []string {
	var docs []string
	docs = append(docs, "date\tRead Avail\tWrite Avail")
	for _, value := range stats.ChartsDocs {
		stat := stats.ServerStatusDoc{}
		for _, doc := range value {
			buf, _ := json.Marshal(doc)
			json.Unmarshal(buf, &stat)
			docs = append(docs, stat.LocalTime.Format("2006-01-02T15:04:05Z")+"\t"+strconv.Itoa(stat.WiredTiger.ConcurrentTransactions.Read.Available)+"\t"+strconv.Itoa(stat.WiredTiger.ConcurrentTransactions.Write.Available))
		}
		break
	}

	return docs
}
