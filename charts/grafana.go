// Copyright 2018 Kuei-chun Chen. All rights reserved.

package charts

import (
	"fmt"
	"sort"
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

var serverStatusChartsLegends = []string{
	"mem_resident", "mem_virtual", "mem_page_faults",
	"conns_available", "conns_current", "conns_created_per_minute",
	"ops_query", "ops_insert", "ops_update", "ops_delete", "ops_getmore", "ops_command",
	"q_active_read", "q_active_write", "q_queued_read", "q_queued_write",
}
var wiredTigerChartsLegends = []string{
	"latency_read", "latency_write", "latency_command",
	"scan_keys", "scan_objects", "scan_sort",
	"wt_cache_max", "wt_cache_used", "wt_cache_dirty",
	"wt_modified_evicted", "wt_unmodified_evicted", "wt_read_in_cache", "wt_written_from_cache",
	"ticket_avail_read", "ticket_avail_write",
}
var systemMetricsChartsLegends = []string{
	"cpu_idle", "cpu_iowait", "cpu_nice", "cpu_softirq", "cpu_steal", "cpu_system", "cpu_user",
	"disks_utils", "disks_iops"}
var replSetChartsLegends = []string{"replication_lags"}

// Grafana simple json data store
// grafana-cli plugins install grafana-simple-json-datasource
type Grafana struct {
	sync.RWMutex
	serverInfo      interface{}
	timeSeriesData  map[string]TimeSeriesDoc
	replicationLags map[string]TimeSeriesDoc
	diskStats       map[string]DiskStats
}

// DiskStats -
type DiskStats struct {
	utilization TimeSeriesDoc
	iops        TimeSeriesDoc
}

var g Grafana

// NewGrafana -
func NewGrafana(d *keyhole.DiagnosticData) *Grafana {
	g = Grafana{serverInfo: d.ServerInfo}
	g.RLock()
	defer g.RUnlock()
	g.ReinitGrafana(d)
	return &g
}

// ReinitGrafana -
func (g *Grafana) ReinitGrafana(d *keyhole.DiagnosticData) {
	btm := time.Now()
	var serverStatusTSD map[string]TimeSeriesDoc
	var wiredTigerTSD map[string]TimeSeriesDoc
	var replicationTSD map[string]TimeSeriesDoc
	var systemMetricsTSD map[string]TimeSeriesDoc
	var replicationLags map[string]TimeSeriesDoc
	var diskStats map[string]DiskStats

	var wg = keyhole.NewWaitGroup(4) // use 4 threads to read
	wg.Add(1)
	go func() {
		defer wg.Done()
		replicationTSD, replicationLags = g.initReplSetGetStatusTimeSeriesDoc(d.ReplSetStatusList) // replSetGetStatus
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		systemMetricsTSD, diskStats = g.initSystemMetricsTimeSeriesDoc(d.SystemMetricsList) // SystemMetrics
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		serverStatusTSD = g.initServerStatusTimeSeriesDoc(d.ServerStatusList) // ServerStatus
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		wiredTigerTSD = g.initWiredTigerTimeSeriesDoc(d.ServerStatusList) // ServerStatus
	}()
	wg.Wait()

	// merge
	g.timeSeriesData = serverStatusTSD
	for k, v := range wiredTigerTSD {
		g.timeSeriesData[k] = v
	}
	g.replicationLags = replicationLags
	for k, v := range replicationTSD {
		g.timeSeriesData[k] = v
	}
	g.diskStats = diskStats
	for k, v := range systemMetricsTSD {
		g.timeSeriesData[k] = v
	}
	etm := time.Now()
	fmt.Println("data points ready, time spent:", etm.Sub(btm).String())
}

func getDataPoint(v float64, t float64) []float64 {
	dp := []float64{}
	dp = append(dp, v)
	dp = append(dp, t)
	return dp
}

func (g *Grafana) initReplSetGetStatusTimeSeriesDoc(replSetGetStatusList []keyhole.ReplSetStatusDoc) (map[string]TimeSeriesDoc, map[string]TimeSeriesDoc) {
	var timeSeriesData = map[string]TimeSeriesDoc{}
	var replicationLags = map[string]TimeSeriesDoc{}
	var hosts []string
	var ts bson.MongoTimestamp

	for _, legend := range replSetChartsLegends {
		timeSeriesData[legend] = TimeSeriesDoc{legend, [][]float64{}}
	}
	for i, stat := range replSetGetStatusList {
		ts = 0
		sort.Slice(stat.Members, func(i, j int) bool { return stat.Members[i].Name < stat.Members[j].Name })
		if i == 0 {
			for n, mb := range stat.Members {
				a := strings.Index(mb.Name, ".")
				b := strings.LastIndex(mb.Name, ":")
				var legend string
				if a < 0 || b < 0 {
					legend = mb.Name
				} else {
					legend = mb.Name[0:a] + mb.Name[b:]
				}
				hosts = append(hosts, legend)
				timeSeriesData[legend] = TimeSeriesDoc{legend, [][]float64{}}
				node := "repl_" + strconv.Itoa(n)
				timeSeriesData[node] = TimeSeriesDoc{node, [][]float64{}}
			}
			continue
		}

		for _, mb := range stat.Members {
			if mb.StateStr == keyhole.PRIMARY {
				ts = keyhole.GetOptime(mb.Optime)
				break
			}
		}

		if ts == 0 {
			continue
		} else {
			t := float64(stat.Date.UnixNano() / 1000 / 1000)
			for i, mb := range stat.Members {
				v := 0.0
				if mb.StateStr == keyhole.SECONDARY {
					v = float64(ts - keyhole.GetOptime(mb.Optime))
				} else if mb.StateStr == keyhole.PRIMARY {
					v = 0
				} else if mb.StateStr == "ARBITER" {
					continue
				}
				x := replicationLags[hosts[i]]
				x.DataPoints = append(x.DataPoints, getDataPoint(v, t))
				replicationLags[hosts[i]] = x
			}
		}
	}

	return timeSeriesData, replicationLags
}

func (g *Grafana) initSystemMetricsTimeSeriesDoc(systemMetricsList []keyhole.SystemMetricsDoc) (map[string]TimeSeriesDoc, map[string]DiskStats) {
	var timeSeriesData = map[string]TimeSeriesDoc{}
	var diskStats = map[string]DiskStats{}
	var pstat = keyhole.SystemMetricsDoc{}

	for _, legend := range systemMetricsChartsLegends {
		timeSeriesData[legend] = TimeSeriesDoc{legend, [][]float64{}}
	}
	for i, stat := range systemMetricsList {
		if i > 0 {
			t := float64(stat.Start.UnixNano() / (1000 * 1000))
			for k, disk := range stat.Disks {
				disk.IO = disk.Reads + disk.Writes
				totalMS := (disk.ReadTimeMS + disk.WriteTimeMS) - (pstat.Disks[k].ReadTimeMS + pstat.Disks[k].WriteTimeMS)
				u := float64(0)
				if totalMS != 0 {
					u = float64(100 * (disk.IOTimeMS - pstat.Disks[k].IOTimeMS) / totalMS)
				}
				if u > 100 {
					continue
				}
				iops := float64(disk.Reads+disk.Writes-(pstat.Disks[k].Reads+pstat.Disks[k].Writes)) / float64(stat.Start.Sub(pstat.Start).Seconds())

				x := diskStats[k]
				x.utilization.DataPoints = append(x.utilization.DataPoints, getDataPoint(u, t))
				x.iops.DataPoints = append(x.iops.DataPoints, getDataPoint(iops, t))
				diskStats[k] = x
			}

			stat.CPU.TotalMS = stat.CPU.IOWaitMS + stat.CPU.IdleMS + stat.CPU.NiceMS + stat.CPU.SoftirqMS + stat.CPU.StealMS + stat.CPU.SystemMS + stat.CPU.UserMS

			x := timeSeriesData["cpu_idle"]
			x.DataPoints = append(x.DataPoints, getDataPoint(100*float64(stat.CPU.IdleMS-pstat.CPU.IdleMS)/float64(stat.CPU.TotalMS-pstat.CPU.TotalMS), t))
			timeSeriesData["cpu_idle"] = x

			x = timeSeriesData["cpu_iowait"]
			x.DataPoints = append(x.DataPoints, getDataPoint(100*float64(stat.CPU.IOWaitMS-pstat.CPU.IOWaitMS)/float64(stat.CPU.TotalMS-pstat.CPU.TotalMS), t))
			timeSeriesData["cpu_iowait"] = x

			x = timeSeriesData["cpu_system"]
			x.DataPoints = append(x.DataPoints, getDataPoint(100*float64(stat.CPU.SystemMS-pstat.CPU.SystemMS)/float64(stat.CPU.TotalMS-pstat.CPU.TotalMS), t))
			timeSeriesData["cpu_system"] = x

			x = timeSeriesData["cpu_user"]
			x.DataPoints = append(x.DataPoints, getDataPoint(100*float64(stat.CPU.UserMS-pstat.CPU.UserMS)/float64(stat.CPU.TotalMS-pstat.CPU.TotalMS), t))
			timeSeriesData["cpu_user"] = x

			x = timeSeriesData["cpu_nice"]
			x.DataPoints = append(x.DataPoints, getDataPoint(100*float64(stat.CPU.NiceMS-pstat.CPU.NiceMS)/float64(stat.CPU.TotalMS-pstat.CPU.TotalMS), t))
			timeSeriesData["cpu_nice"] = x

			x = timeSeriesData["cpu_steal"]
			x.DataPoints = append(x.DataPoints, getDataPoint(100*float64(stat.CPU.StealMS-pstat.CPU.StealMS)/float64(stat.CPU.TotalMS-pstat.CPU.TotalMS), t))
			timeSeriesData["cpu_steal"] = x

			x = timeSeriesData["cpu_softirq"]
			x.DataPoints = append(x.DataPoints, getDataPoint(100*float64(stat.CPU.SoftirqMS-pstat.CPU.SoftirqMS)/float64(stat.CPU.TotalMS-pstat.CPU.TotalMS), t))
			timeSeriesData["cpu_softirq"] = x
		}

		pstat = stat
	}
	return timeSeriesData, diskStats
}

func (g *Grafana) initServerStatusTimeSeriesDoc(serverStatusList []keyhole.ServerStatusDoc) map[string]TimeSeriesDoc {
	var timeSeriesData = map[string]TimeSeriesDoc{}
	pstat := keyhole.ServerStatusDoc{}
	var x TimeSeriesDoc

	for _, legend := range serverStatusChartsLegends {
		timeSeriesData[legend] = TimeSeriesDoc{legend, [][]float64{}}
	}
	for i, stat := range serverStatusList {
		if stat.Uptime > pstat.Uptime {
			t := float64(stat.LocalTime.UnixNano() / (1000 * 1000))

			x = timeSeriesData["mem_resident"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Mem.Resident)/1024, t))
			timeSeriesData["mem_resident"] = x

			x = timeSeriesData["mem_virtual"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Mem.Virtual)/1024, t))
			timeSeriesData["mem_virtual"] = x

			x = timeSeriesData["conns_available"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Connections.Available), t))
			timeSeriesData["conns_available"] = x

			x = timeSeriesData["conns_current"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Connections.Current), t))
			timeSeriesData["conns_current"] = x

			x = timeSeriesData["q_active_read"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.GlobalLock.ActiveClients.Readers), t))
			timeSeriesData["q_active_read"] = x

			x = timeSeriesData["q_active_write"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.GlobalLock.ActiveClients.Writers), t))
			timeSeriesData["q_active_write"] = x

			x = timeSeriesData["q_queued_read"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.GlobalLock.CurrentQueue.Readers), t))
			timeSeriesData["q_queued_read"] = x

			x = timeSeriesData["q_queued_write"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.GlobalLock.CurrentQueue.Writers), t))
			timeSeriesData["q_queued_write"] = x

			if i > 0 {
				minutes := stat.LocalTime.Sub(pstat.LocalTime).Minutes()

				x = timeSeriesData["mem_page_faults"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.ExtraInfo.PageFaults-pstat.ExtraInfo.PageFaults), t))
				timeSeriesData["mem_page_faults"] = x

				x = timeSeriesData["conns_created_per_minute"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Connections.TotalCreated-pstat.Connections.TotalCreated)/minutes, t))
				timeSeriesData["conns_created_per_minute"] = x

				x = timeSeriesData["ops_query"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.OpCounters.Query-pstat.OpCounters.Query), t))
				timeSeriesData["ops_query"] = x

				x = timeSeriesData["ops_insert"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.OpCounters.Insert-pstat.OpCounters.Insert), t))
				timeSeriesData["ops_insert"] = x

				x = timeSeriesData["ops_update"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.OpCounters.Update-pstat.OpCounters.Update), t))
				timeSeriesData["ops_update"] = x

				x = timeSeriesData["ops_delete"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.OpCounters.Delete-pstat.OpCounters.Delete), t))
				timeSeriesData["ops_delete"] = x

				x = timeSeriesData["ops_getmore"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.OpCounters.Getmore-pstat.OpCounters.Getmore), t))
				timeSeriesData["ops_getmore"] = x

				x = timeSeriesData["ops_command"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.OpCounters.Command-pstat.OpCounters.Command), t))
				timeSeriesData["ops_command"] = x
			} // if i > 0
		} // if stat.Uptime > pstat.Uptime

		pstat = stat
	}
	return timeSeriesData
}

func (g *Grafana) initWiredTigerTimeSeriesDoc(serverStatusList []keyhole.ServerStatusDoc) map[string]TimeSeriesDoc {
	var timeSeriesData = map[string]TimeSeriesDoc{}
	pstat := keyhole.ServerStatusDoc{}
	var x TimeSeriesDoc

	for _, legend := range wiredTigerChartsLegends {
		timeSeriesData[legend] = TimeSeriesDoc{legend, [][]float64{}}
	}
	for i, stat := range serverStatusList {
		if stat.Uptime > pstat.Uptime {
			t := float64(stat.LocalTime.UnixNano() / (1000 * 1000))

			x = timeSeriesData["wt_cache_max"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.Cache.MaxBytesConfigured)/(1024*1024*1024), t))
			timeSeriesData["wt_cache_max"] = x

			x = timeSeriesData["wt_cache_used"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.Cache.CurrentlyInCache)/(1024*1024*1024), t))
			timeSeriesData["wt_cache_used"] = x

			x = timeSeriesData["wt_cache_dirty"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.Cache.TrackedDirtyBytes)/(1024*1024*1024), t))
			timeSeriesData["wt_cache_dirty"] = x

			x = timeSeriesData["ticket_avail_read"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.ConcurrentTransactions.Read.Available), t))
			timeSeriesData["ticket_avail_read"] = x

			x = timeSeriesData["ticket_avail_write"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.ConcurrentTransactions.Write.Available), t))
			timeSeriesData["ticket_avail_write"] = x

			r := 0.0
			if stat.OpLatencies.Reads.Ops > 0 {
				r = float64(stat.OpLatencies.Reads.Latency) / float64(stat.OpLatencies.Reads.Ops) / 1000
			}
			x = timeSeriesData["latency_read"]
			x.DataPoints = append(x.DataPoints, getDataPoint(r, t))
			timeSeriesData["latency_read"] = x

			w := 0.0
			if stat.OpLatencies.Writes.Ops > 0 {
				w = float64(stat.OpLatencies.Writes.Latency) / float64(stat.OpLatencies.Writes.Ops) / 1000
			}
			x = timeSeriesData["latency_write"]
			x.DataPoints = append(x.DataPoints, getDataPoint(w, t))
			timeSeriesData["latency_write"] = x

			c := 0.0
			if stat.OpLatencies.Commands.Ops > 0 {
				c = float64(stat.OpLatencies.Commands.Latency) / float64(stat.OpLatencies.Commands.Ops) / 1000
			}
			x = timeSeriesData["latency_command"]
			x.DataPoints = append(x.DataPoints, getDataPoint(c, t))
			timeSeriesData["latency_command"] = x

			if i > 0 {
				minutes := stat.LocalTime.Sub(pstat.LocalTime).Minutes()

				x = timeSeriesData["scan_keys"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Metrics.QueryExecutor.Scanned-pstat.Metrics.QueryExecutor.Scanned), t))
				timeSeriesData["scan_keys"] = x

				x = timeSeriesData["scan_objects"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Metrics.QueryExecutor.ScannedObjects-pstat.Metrics.QueryExecutor.ScannedObjects), t))
				timeSeriesData["scan_objects"] = x

				x = timeSeriesData["scan_sort"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Metrics.Operation.ScanAndOrder-pstat.Metrics.Operation.ScanAndOrder), t))
				timeSeriesData["scan_sort"] = x

				x = timeSeriesData["wt_modified_evicted"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.Cache.ModifiedPagesEvicted-pstat.WiredTiger.Cache.ModifiedPagesEvicted)/minutes, t))
				timeSeriesData["wt_modified_evicted"] = x

				x = timeSeriesData["wt_unmodified_evicted"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.Cache.UnmodifiedPagesEvicted-pstat.WiredTiger.Cache.UnmodifiedPagesEvicted)/minutes, t))
				timeSeriesData["wt_unmodified_evicted"] = x

				x = timeSeriesData["wt_read_in_cache"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.Cache.PagesReadIntoCache-pstat.WiredTiger.Cache.PagesReadIntoCache)/minutes, t))
				timeSeriesData["wt_read_in_cache"] = x

				x = timeSeriesData["wt_written_from_cache"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.Cache.PagesWrittenFromCache-pstat.WiredTiger.Cache.PagesWrittenFromCache)/minutes, t))
				timeSeriesData["wt_written_from_cache"] = x
			} // if i > 0
		} // if stat.Uptime > pstat.Uptime

		pstat = stat
	}
	return timeSeriesData
}
