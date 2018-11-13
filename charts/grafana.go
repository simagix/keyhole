// Copyright 2018 Kuei-chun Chen. All rights reserved.

package charts

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
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

var chartsLegends = []string{
	"mem_resident", "mem_virtual", "mem_page_faults",
	"conns_available", "conns_current", "conns_created_per_minute",
	"ops_query", "ops_insert", "ops_update", "ops_delete", "ops_getmore", "ops_command",
	"q_active_read", "q_active_write", "q_queued_read", "q_queued_write",
	"latency_read", "latency_write", "latency_command",
	"scan_keys", "scan_objects", "scan_sort",
	"wt_cache_max", "wt_cache_used", "wt_cache_dirty",
	"wt_modified_evicted", "wt_unmodified_evicted", "wt_read_in_cache", "wt_written_from_cache",
	"ticket_avail_read", "ticket_avail_write",
	"cpu_idle", "cpu_iowait", "cpu_nice", "cpu_softirq", "cpu_steal", "cpu_system", "cpu_user",
	"disks_utils",
	"replication_lags",
}

// Grafana simple json data store
// grafana-cli plugins install grafana-simple-json-datasource
type Grafana struct {
	sync.RWMutex
	timeSeriesData  map[string]TimeSeriesDoc
	replicationLags map[string]TimeSeriesDoc
	diskUtils       map[string]TimeSeriesDoc
}

// NewGrafana -
func NewGrafana(ChartsDocs map[string][]bson.M) *Grafana {
	g := Grafana{
		timeSeriesData:  map[string]TimeSeriesDoc{},
		replicationLags: map[string]TimeSeriesDoc{},
		diskUtils:       map[string]TimeSeriesDoc{},
	}
	g.RLock()
	defer g.RUnlock()
	for _, legend := range chartsLegends {
		g.timeSeriesData[legend] = TimeSeriesDoc{legend, [][]float64{}}
	}
	g.initServerStatusTimeSeriesDoc(ChartsDocs["serverStatus"])   // ServerStatus
	g.initSystemMetricsTimeSeriesDoc(ChartsDocs["systemMetrics"]) // SystemMetrics
	g.initReplSetGetStatusTimeSeriesDoc(GetReplLagsTSV())         // replSetGetStatus
	return &g
}

func getDataPoint(v float64, t float64) []float64 {
	dp := []float64{}
	dp = append(dp, v)
	dp = append(dp, t)
	return dp
}

func (g *Grafana) initReplSetGetStatusTimeSeriesDoc(replSetGetStatusTSV []string) {
	var hosts []string
	for i, line := range replSetGetStatusTSV {
		if i == 0 {
			hosts = strings.Split(line, "\t")
			for n, legend := range hosts {
				if n > 0 {
					g.timeSeriesData[legend] = TimeSeriesDoc{legend, [][]float64{}}
					node := "repl_" + strconv.Itoa(n)
					g.timeSeriesData[node] = TimeSeriesDoc{node, [][]float64{}}
				}
			}
			continue
		}

		tokens := strings.Split(line, "\t")
		t1, _ := time.Parse(time.RFC3339, tokens[0])
		t := float64(t1.UnixNano() / (1000 * 1000))

		for i, token := range tokens {
			if i == 0 {
				continue
			}

			v, _ := strconv.ParseFloat(token, 64)
			x := g.replicationLags[hosts[i]]
			x.DataPoints = append(x.DataPoints, getDataPoint(v, t))
			g.replicationLags[hosts[i]] = x
		}
	}
}

func (g *Grafana) initSystemMetricsTimeSeriesDoc(systemMetricsList []bson.M) {
	var stat = keyhole.SystemMetricsDoc{}
	var pstat = keyhole.SystemMetricsDoc{}
	var pdstat = keyhole.SystemMetricsDoc{}
	var disk = keyhole.DiskMetrics{}
	var pmdisk = map[string]keyhole.DiskMetrics{}

	for i, doc := range systemMetricsList {
		buf, _ := json.Marshal(doc)
		json.Unmarshal(buf, &stat)

		t := float64(stat.Start.UnixNano() / (1000 * 1000))
		var isMatch = false

		// althought data points come in every 5 minutes but disks metrics have
		// many duplicate points.  All data points from the same file has same values
		// conclusion is to drop the chart
		for k, v := range stat.Disks {
			b, _ := json.Marshal(v)
			json.Unmarshal(b, &disk)
			disk.IO = disk.Reads + disk.Writes

			if i == 0 {
				pmdisk[k] = disk
				pdstat = stat
			} else if disk.IO != pmdisk[k].IO {
				u := float64(disk.IO-pmdisk[k].IO) / stat.Start.Sub(pdstat.Start).Seconds() // IOPS
				// u := 100 * float64(disk.IOTimeMS-pmdisk[k].IOTimeMS) / (stat.Start.Sub(pdstat.Start).Seconds() * 1000) // Disk Utilization (%)
				x := g.diskUtils[k]
				x.DataPoints = append(x.DataPoints, getDataPoint(u, t))
				g.diskUtils[k] = x
				pmdisk[k] = disk
				isMatch = true
			}
		}
		if isMatch {
			pdstat = stat
		}
		stat.CPU.TotalMS = stat.CPU.IOWaitMS + stat.CPU.IdleMS + stat.CPU.NiceMS + stat.CPU.SoftirqMS + stat.CPU.StealMS + stat.CPU.SystemMS + stat.CPU.UserMS

		if i > 0 && stat.CPU.TotalMS != 0 {
			x := g.timeSeriesData["cpu_idle"]
			x.DataPoints = append(x.DataPoints, getDataPoint(100*float64(stat.CPU.IdleMS-pstat.CPU.IdleMS)/float64(stat.CPU.TotalMS-pstat.CPU.TotalMS), t))
			g.timeSeriesData["cpu_idle"] = x

			x = g.timeSeriesData["cpu_iowait"]
			x.DataPoints = append(x.DataPoints, getDataPoint(100*float64(stat.CPU.IOWaitMS-pstat.CPU.IOWaitMS)/float64(stat.CPU.TotalMS-pstat.CPU.TotalMS), t))
			g.timeSeriesData["cpu_iowait"] = x

			x = g.timeSeriesData["cpu_system"]
			x.DataPoints = append(x.DataPoints, getDataPoint(100*float64(stat.CPU.SystemMS-pstat.CPU.SystemMS)/float64(stat.CPU.TotalMS-pstat.CPU.TotalMS), t))
			g.timeSeriesData["cpu_system"] = x

			x = g.timeSeriesData["cpu_user"]
			x.DataPoints = append(x.DataPoints, getDataPoint(100*float64(stat.CPU.UserMS-pstat.CPU.UserMS)/float64(stat.CPU.TotalMS-pstat.CPU.TotalMS), t))
			g.timeSeriesData["cpu_user"] = x

			x = g.timeSeriesData["cpu_nice"]
			x.DataPoints = append(x.DataPoints, getDataPoint(100*float64(stat.CPU.NiceMS-pstat.CPU.NiceMS)/float64(stat.CPU.TotalMS-pstat.CPU.TotalMS), t))
			g.timeSeriesData["cpu_nice"] = x

			x = g.timeSeriesData["cpu_steal"]
			x.DataPoints = append(x.DataPoints, getDataPoint(100*float64(stat.CPU.StealMS-pstat.CPU.StealMS)/float64(stat.CPU.TotalMS-pstat.CPU.TotalMS), t))
			g.timeSeriesData["cpu_steal"] = x

			x = g.timeSeriesData["cpu_softirq"]
			x.DataPoints = append(x.DataPoints, getDataPoint(100*float64(stat.CPU.SoftirqMS-pstat.CPU.SoftirqMS)/float64(stat.CPU.TotalMS-pstat.CPU.TotalMS), t))
			g.timeSeriesData["cpu_softirq"] = x
		}

		pstat = stat
	}
}

func (g *Grafana) initServerStatusTimeSeriesDoc(serverStatusList []bson.M) {
	pstat := keyhole.ServerStatusDoc{}
	stat := keyhole.ServerStatusDoc{}
	var x TimeSeriesDoc

	for i, doc := range ChartsDocs["serverStatus"] {
		buf, _ := json.Marshal(doc)
		json.Unmarshal(buf, &stat)
		if stat.Uptime > pstat.Uptime {
			t := float64(stat.LocalTime.UnixNano() / (1000 * 1000))

			x = g.timeSeriesData["mem_resident"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Mem.Resident)/1024, t))
			g.timeSeriesData["mem_resident"] = x

			x = g.timeSeriesData["mem_virtual"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Mem.Virtual)/1024, t))
			g.timeSeriesData["mem_virtual"] = x

			x = g.timeSeriesData["conns_available"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Connections.Available), t))
			g.timeSeriesData["conns_available"] = x

			x = g.timeSeriesData["conns_current"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Connections.Current), t))
			g.timeSeriesData["conns_current"] = x

			x = g.timeSeriesData["q_active_read"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.GlobalLock.ActiveClients.Readers), t))
			g.timeSeriesData["q_active_read"] = x

			x = g.timeSeriesData["q_active_write"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.GlobalLock.ActiveClients.Writers), t))
			g.timeSeriesData["q_active_write"] = x

			x = g.timeSeriesData["q_queued_read"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.GlobalLock.CurrentQueue.Readers), t))
			g.timeSeriesData["q_queued_read"] = x

			x = g.timeSeriesData["q_queued_write"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.GlobalLock.CurrentQueue.Writers), t))
			g.timeSeriesData["q_queued_write"] = x

			r := 0.0
			if stat.OpLatencies.Reads.Ops > 0 {
				r = float64(stat.OpLatencies.Reads.Latency) / float64(stat.OpLatencies.Reads.Ops) / 1000
			}
			x = g.timeSeriesData["latency_read"]
			x.DataPoints = append(x.DataPoints, getDataPoint(r, t))
			g.timeSeriesData["latency_read"] = x

			w := 0.0
			if stat.OpLatencies.Writes.Ops > 0 {
				w = float64(stat.OpLatencies.Writes.Latency) / float64(stat.OpLatencies.Writes.Ops) / 1000
			}
			x = g.timeSeriesData["latency_write"]
			x.DataPoints = append(x.DataPoints, getDataPoint(w, t))
			g.timeSeriesData["latency_write"] = x

			c := 0.0
			if stat.OpLatencies.Commands.Ops > 0 {
				c = float64(stat.OpLatencies.Commands.Latency) / float64(stat.OpLatencies.Commands.Ops) / 1000
			}
			x = g.timeSeriesData["latency_command"]
			x.DataPoints = append(x.DataPoints, getDataPoint(c, t))
			g.timeSeriesData["latency_command"] = x

			x = g.timeSeriesData["wt_cache_max"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.Cache.MaxBytesConfigured)/(1024*1024*1024), t))
			g.timeSeriesData["wt_cache_max"] = x

			x = g.timeSeriesData["wt_cache_used"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.Cache.CurrentlyInCache)/(1024*1024*1024), t))
			g.timeSeriesData["wt_cache_used"] = x

			x = g.timeSeriesData["wt_cache_dirty"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.Cache.TrackedDirtyBytes)/(1024*1024*1024), t))
			g.timeSeriesData["wt_cache_dirty"] = x

			x = g.timeSeriesData["ticket_avail_read"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.ConcurrentTransactions.Read.Available), t))
			g.timeSeriesData["ticket_avail_read"] = x

			x = g.timeSeriesData["ticket_avail_write"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.ConcurrentTransactions.Write.Available), t))
			g.timeSeriesData["ticket_avail_write"] = x

			if i > 0 {
				x = g.timeSeriesData["mem_page_faults"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.ExtraInfo.PageFaults-pstat.ExtraInfo.PageFaults), t))
				g.timeSeriesData["mem_page_faults"] = x

				minutes := stat.LocalTime.Sub(pstat.LocalTime).Minutes()
				x = g.timeSeriesData["conns_created_per_minute"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Connections.TotalCreated-pstat.Connections.TotalCreated)/minutes, t))
				g.timeSeriesData["conns_created_per_minute"] = x

				x = g.timeSeriesData["ops_query"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.OpCounters.Query-pstat.OpCounters.Query), t))
				g.timeSeriesData["ops_query"] = x

				x = g.timeSeriesData["ops_insert"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.OpCounters.Insert-pstat.OpCounters.Insert), t))
				g.timeSeriesData["ops_insert"] = x

				x = g.timeSeriesData["ops_update"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.OpCounters.Update-pstat.OpCounters.Update), t))
				g.timeSeriesData["ops_update"] = x

				x = g.timeSeriesData["ops_delete"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.OpCounters.Delete-pstat.OpCounters.Delete), t))
				g.timeSeriesData["ops_delete"] = x

				x = g.timeSeriesData["ops_getmore"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.OpCounters.Getmore-pstat.OpCounters.Getmore), t))
				g.timeSeriesData["ops_getmore"] = x

				x = g.timeSeriesData["ops_command"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.OpCounters.Command-pstat.OpCounters.Command), t))
				g.timeSeriesData["ops_command"] = x

				x = g.timeSeriesData["scan_keys"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Metrics.QueryExecutor.Scanned-pstat.Metrics.QueryExecutor.Scanned), t))
				g.timeSeriesData["scan_keys"] = x

				x = g.timeSeriesData["scan_objects"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Metrics.QueryExecutor.ScannedObjects-pstat.Metrics.QueryExecutor.ScannedObjects), t))
				g.timeSeriesData["scan_objects"] = x

				x = g.timeSeriesData["scan_sort"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Metrics.Operation.ScanAndOrder-pstat.Metrics.Operation.ScanAndOrder), t))
				g.timeSeriesData["scan_sort"] = x

				x = g.timeSeriesData["wt_modified_evicted"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.Cache.ModifiedPagesEvicted-pstat.WiredTiger.Cache.ModifiedPagesEvicted)/minutes, t))
				g.timeSeriesData["wt_modified_evicted"] = x

				x = g.timeSeriesData["wt_unmodified_evicted"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.Cache.UnmodifiedPagesEvicted-pstat.WiredTiger.Cache.UnmodifiedPagesEvicted)/minutes, t))
				g.timeSeriesData["wt_unmodified_evicted"] = x

				x = g.timeSeriesData["wt_read_in_cache"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.Cache.PagesReadIntoCache-pstat.WiredTiger.Cache.PagesReadIntoCache)/minutes, t))
				g.timeSeriesData["wt_read_in_cache"] = x

				x = g.timeSeriesData["wt_written_from_cache"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.Cache.PagesWrittenFromCache-pstat.WiredTiger.Cache.PagesWrittenFromCache)/minutes, t))
				g.timeSeriesData["wt_written_from_cache"] = x
			} // if i > 0
		} // if stat.Uptime > pstat.Uptime

		pstat = stat
	}
}

func (g *Grafana) handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path[1:] == "grafana/" {
		fmt.Fprintf(w, "ok\n")
	} else if r.URL.Path[1:] == "grafana/query" {
		g.query(w, r)
	} else if r.URL.Path[1:] == "grafana/search" {
		g.search(w, r)
	}
}

func (g *Grafana) search(w http.ResponseWriter, r *http.Request) {
	var list []string

	for _, doc := range g.timeSeriesData {
		list = append(list, doc.Target)
	}

	json.NewEncoder(w).Encode(list)
}

func (g *Grafana) query(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var qr QueryRequest
	err := decoder.Decode(&qr)
	if err != nil {
		return
	}

	var tsData []TimeSeriesDoc
	for _, target := range qr.Targets {
		if target.Target == "replication_lags" { // replaced with actual hostname
			for k, v := range g.replicationLags {
				data := v
				data.Target = k
				tsData = append(tsData, data)
			}
		} else if target.Target == "disks_utils" {
			for k, v := range g.diskUtils {
				data := v
				data.Target = k
				tsData = append(tsData, data)
			}
		} else {
			tsData = append(tsData, g.timeSeriesData[target.Target])
		}
	}

	json.NewEncoder(w).Encode(tsData)
}
