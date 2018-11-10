package charts

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/globalsign/mgo/bson"
	keyhole "github.com/simagix/keyhole/core"
)

// TimeSeriesDoc -
type TimeSeriesDoc struct {
	Target     string      `json:"target"`
	DataPoints [][]float64 `json:"datapoints"`
}

// RangeDoc -
type RangeDoc struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// TargetDoc -
type TargetDoc struct {
	Target string `json:"target"`
	RefID  string `json:"refId"`
	Type   string `json:"type"`
}

// QueryRequest -
type QueryRequest struct {
	Timezone string      `json:"timezone"`
	Range    RangeDoc    `json:"range"`
	Targets  []TargetDoc `json:"targets"`
}

var timeSeriesData map[string]TimeSeriesDoc
var chartsLegends = []string{"mem_resident", "mem_virtual", "mem_page_faults",
	"conns_available", "conns_current", "conns_created_per_minute",
	"ops_query", "ops_insert", "ops_update", "ops_delete", "ops_getmore", "ops_command",
	"q_active_read", "q_active_write", "q_queued_read", "q_queued_write",
	"latency_read", "latency_write", "latency_command",
	"scan_keys", "scan_objects", "scan_sort",
	"wt_cache_max", "wt_cache_used", "wt_cache_dirty",
	"wt_modified_evicted", "wt_unmodified_evicted", "wt_read_in_cache", "wt_written_from_cache",
	"ticket_avail_read", "ticket_avail_write"}

// InitGrafana -
func InitGrafana() {
	timeSeriesData = getTimeSeriesDoc(keyhole.ChartsDocs["serverStatus"], keyhole.ChartsDocs["replSetGetStatus"])
}

func getDataPoint(v float64, t float64) []float64 {
	dp := []float64{}
	dp = append(dp, v)
	dp = append(dp, t)
	return dp
}

func getTimeSeriesDoc(serverStatusList []bson.M, replSetGetStatus []bson.M) map[string]TimeSeriesDoc {
	var tsMap = map[string]TimeSeriesDoc{}
	pstat := keyhole.ServerStatusDoc{}
	stat := keyhole.ServerStatusDoc{}
	var x TimeSeriesDoc

	for _, legend := range chartsLegends {
		tsMap[legend] = TimeSeriesDoc{legend, [][]float64{}}
	}

	for i, doc := range keyhole.ChartsDocs["serverStatus"] {
		buf, _ := json.Marshal(doc)
		json.Unmarshal(buf, &stat)
		if stat.Uptime > pstat.Uptime {
			t := float64(stat.LocalTime.UnixNano() / (1000 * 1000))

			x = tsMap["mem_resident"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Mem.Resident)/1024, t))
			tsMap["mem_resident"] = x

			x = tsMap["mem_virtual"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Mem.Virtual)/1024, t))
			tsMap["mem_virtual"] = x

			x = tsMap["conns_available"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Connections.Available), t))
			tsMap["conns_available"] = x

			x = tsMap["conns_current"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Connections.Current), t))
			tsMap["conns_current"] = x

			x = tsMap["q_active_read"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.GlobalLock.ActiveClients.Readers), t))
			tsMap["q_active_read"] = x

			x = tsMap["q_active_write"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.GlobalLock.ActiveClients.Writers), t))
			tsMap["q_active_write"] = x

			x = tsMap["q_queued_read"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.GlobalLock.CurrentQueue.Readers), t))
			tsMap["q_queued_read"] = x

			x = tsMap["q_queued_write"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.GlobalLock.CurrentQueue.Writers), t))
			tsMap["q_queued_write"] = x

			r := 0.0
			if stat.OpLatencies.Reads.Ops > 0 {
				r = float64(stat.OpLatencies.Reads.Latency) / float64(stat.OpLatencies.Reads.Ops) / 1000
			}
			x = tsMap["latency_read"]
			x.DataPoints = append(x.DataPoints, getDataPoint(r, t))
			tsMap["latency_read"] = x

			w := 0.0
			if stat.OpLatencies.Writes.Ops > 0 {
				w = float64(stat.OpLatencies.Writes.Latency) / float64(stat.OpLatencies.Writes.Ops) / 1000
			}
			x = tsMap["latency_write"]
			x.DataPoints = append(x.DataPoints, getDataPoint(w, t))
			tsMap["latency_write"] = x

			c := 0.0
			if stat.OpLatencies.Commands.Ops > 0 {
				c = float64(stat.OpLatencies.Commands.Latency) / float64(stat.OpLatencies.Commands.Ops) / 1000
			}
			x = tsMap["latency_command"]
			x.DataPoints = append(x.DataPoints, getDataPoint(c, t))
			tsMap["latency_command"] = x

			x = tsMap["wt_cache_max"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.Cache.MaxBytesConfigured)/(1024*1024*1024), t))
			tsMap["wt_cache_max"] = x

			x = tsMap["wt_cache_used"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.Cache.CurrentlyInCache)/(1024*1024*1024), t))
			tsMap["wt_cache_used"] = x

			x = tsMap["wt_cache_dirty"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.Cache.TrackedDirtyBytes)/(1024*1024*1024), t))
			tsMap["wt_cache_dirty"] = x

			x = tsMap["ticket_avail_read"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.ConcurrentTransactions.Read.Available), t))
			tsMap["ticket_avail_read"] = x

			x = tsMap["ticket_avail_write"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.ConcurrentTransactions.Write.Available), t))
			tsMap["ticket_avail_write"] = x

			if i > 0 {
				x = tsMap["mem_page_faults"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.ExtraInfo.PageFaults-pstat.ExtraInfo.PageFaults), t))
				tsMap["mem_page_faults"] = x

				minutes := stat.LocalTime.Sub(pstat.LocalTime).Minutes()
				x = tsMap["conns_created_per_minute"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Connections.TotalCreated-pstat.Connections.TotalCreated)/minutes, t))
				tsMap["conns_created_per_minute"] = x

				x = tsMap["ops_query"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.OpCounters.Query-pstat.OpCounters.Query), t))
				tsMap["ops_query"] = x

				x = tsMap["ops_insert"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.OpCounters.Insert-pstat.OpCounters.Insert), t))
				tsMap["ops_insert"] = x

				x = tsMap["ops_update"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.OpCounters.Update-pstat.OpCounters.Update), t))
				tsMap["ops_update"] = x

				x = tsMap["ops_delete"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.OpCounters.Delete-pstat.OpCounters.Delete), t))
				tsMap["ops_delete"] = x

				x = tsMap["ops_getmore"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.OpCounters.Getmore-pstat.OpCounters.Getmore), t))
				tsMap["ops_getmore"] = x

				x = tsMap["ops_command"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.OpCounters.Command-pstat.OpCounters.Command), t))
				tsMap["ops_command"] = x

				x = tsMap["scan_keys"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Metrics.QueryExecutor.Scanned-pstat.Metrics.QueryExecutor.Scanned), t))
				tsMap["scan_keys"] = x

				x = tsMap["scan_objects"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Metrics.QueryExecutor.ScannedObjects-pstat.Metrics.QueryExecutor.ScannedObjects), t))
				tsMap["scan_objects"] = x

				x = tsMap["scan_sort"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Metrics.Operation.ScanAndOrder-pstat.Metrics.Operation.ScanAndOrder), t))
				tsMap["scan_sort"] = x

				x = tsMap["wt_modified_evicted"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.Cache.ModifiedPagesEvicted-pstat.WiredTiger.Cache.ModifiedPagesEvicted)/minutes, t))
				tsMap["wt_modified_evicted"] = x

				x = tsMap["wt_unmodified_evicted"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.Cache.UnmodifiedPagesEvicted-pstat.WiredTiger.Cache.UnmodifiedPagesEvicted)/minutes, t))
				tsMap["wt_unmodified_evicted"] = x

				x = tsMap["wt_read_in_cache"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.Cache.PagesReadIntoCache-pstat.WiredTiger.Cache.PagesReadIntoCache)/minutes, t))
				tsMap["wt_read_in_cache"] = x

				x = tsMap["wt_written_from_cache"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.Cache.PagesWrittenFromCache-pstat.WiredTiger.Cache.PagesWrittenFromCache)/minutes, t))
				tsMap["wt_written_from_cache"] = x
			} // if i > 0
		} // if stat.Uptime > pstat.Uptime

		pstat = stat
	}

	return tsMap
}

// grafana-cli plugins install grafana-simple-json-datasource
func grafana(w http.ResponseWriter, r *http.Request) {
	var str string
	if r.URL.Path[1:] == "grafana/" {
		fmt.Fprintf(w, "ok\n")

	} else if r.URL.Path[1:] == "grafana/query" {
		query(w, r)

	} else if r.URL.Path[1:] == "grafana/search" {
		search(w, r)

	} else {
		str = "TODO: " + r.URL.Path + "\n"
		fmt.Fprintf(w, str)
	}
}

func search(w http.ResponseWriter, r *http.Request) {
	var list []string

	for _, doc := range timeSeriesData {
		list = append(list, doc.Target)
	}

	json.NewEncoder(w).Encode(list)
}

func query(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var qr QueryRequest
	err := decoder.Decode(&qr)
	if err != nil {
		return
	}

	var tsData []TimeSeriesDoc
	for _, target := range qr.Targets {
		tsData = append(tsData, timeSeriesData[target.Target])
	}

	json.NewEncoder(w).Encode(tsData)
}
