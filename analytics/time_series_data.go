// Copyright 2019 Kuei-chun Chen. All rights reserved.

package analytics

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

var mb = (1024.0 * 1024)

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
	"conns_active", "conns_available", "conns_current", "conns_created/s",
	"latency_read", "latency_write", "latency_command",
	"net_in", "net_out", "net_requests", "net_physical_in", "net_physical_out",
	"ops_query", "ops_insert", "ops_update", "ops_delete", "ops_getmore", "ops_command",
	"q_active_read", "q_active_write", "q_queued_read", "q_queued_write",
	"scan_keys", "scan_objects", "scan_sort", "tcmalloc_current_allocated_bytes",
	"tcmalloc_heap_size", "tcmalloc_pageheap_free_bytes", "tcmalloc_pageheap_unmapped_bytes",
	"tcmalloc_max_total_thread_cache_bytes", "tcmalloc_current_total_thread_cache_bytes",
	"tcmalloc_total_free_bytes", "tcmalloc_central_cache_free_bytes", "tcmalloc_transfer_cache_free_bytes",
	"tcmalloc_thread_cache_free_bytes", "tcmalloc_aggressive_memory_decommit", "tcmalloc_pageheap_committed_bytes",
	"tcmalloc_pageheap_scavenge_count", "tcmalloc_pageheap_commit_count", "tcmalloc_pageheap_total_commit_bytes",
	"tcmalloc_pageheap_decommit_count", "tcmalloc_pageheap_total_decommit_bytes", "tcmalloc_pageheap_reserve_count",
	"tcmalloc_pageheap_total_reserve_bytes", "tcmalloc_spinlock_total_delay_ns", "tcmalloc_release_rate",
}
var wiredTigerChartsLegends = []string{
	"wt_blkmgr_read", "wt_blkmgr_written", "wt_blkmgr_written_checkpoint",
	"wt_cache_max", "wt_cache_used", "wt_cache_dirty",
	"wt_modified_evicted", "wt_unmodified_evicted", "wt_cache_read_in", "wt_cache_written_from",
	"wt_dhandles_active", "ticket_avail_read", "ticket_avail_write",
}
var systemMetricsChartsLegends = []string{
	"cpu_idle", "cpu_iowait", "cpu_nice", "cpu_softirq", "cpu_steal", "cpu_system", "cpu_user",
	"disks_utils", "disks_iops", "io_in_progress", "read_time_ms", "write_time_ms", "io_queued_ms"}
var replSetChartsLegends = []string{"replication_lags"}

func getDataPoint(v float64, t float64) []float64 {
	dp := []float64{}
	if v < 0 {
		v = 0
	}
	dp = append(dp, v)
	dp = append(dp, t)
	return dp
}

func getReplSetGetStatusTimeSeriesDoc(replSetGetStatusList []ReplSetStatusDoc, legends *[]string) (map[string]TimeSeriesDoc, map[string]TimeSeriesDoc) {
	var timeSeriesData = map[string]TimeSeriesDoc{}
	var replicationLags = map[string]TimeSeriesDoc{}
	var hosts []string
	var ts int64

	for _, legend := range replSetChartsLegends {
		timeSeriesData[legend] = TimeSeriesDoc{legend, [][]float64{}}
	}
	for _, stat := range replSetGetStatusList {
		if len(stat.Members) == 0 { // missing, shouldn't happen
			continue
		}
		ts = 0
		sort.Slice(stat.Members, func(i, j int) bool { return stat.Members[i].Name < stat.Members[j].Name })
		if len(hosts) == 0 || len(hosts) != len(stat.Members) {
			hosts = hosts[:0]
			for n, mb := range stat.Members {
				hostname := fmt.Sprintf("host-%v", n)
				a := strings.Index(mb.Name, ".")
				b := strings.LastIndex(mb.Name, ":")
				var legend string
				if a < 0 || b < 0 {
					legend = mb.Name
				} else {
					legend = mb.Name[0:a] + mb.Name[b:]
				}
				if len(*legends) == 0 {
					log.Println(hostname, legend)
				}
				hosts = append(hosts, hostname)
				timeSeriesData[legend] = TimeSeriesDoc{legend, [][]float64{}}
				node := "repl_" + strconv.Itoa(n)
				timeSeriesData[node] = TimeSeriesDoc{node, [][]float64{}}
			}
			continue
		}

		for _, mb := range stat.Members {
			if mb.State == 1 {
				ts = GetOptime(mb.Optime)
				break
			}
		}

		if ts == 0 {
			continue
		} else {
			t := float64(stat.Date.UnixNano() / 1000 / 1000)
			for i, mb := range stat.Members {
				v := 0.0
				if mb.State == 2 { // SECONDARY
					v = float64(ts - GetOptime(mb.Optime))
				} else if mb.State == 1 { // PRIMARY
					v = 0
				} else if mb.State == 7 { // ARBITER
					continue
				}
				x := replicationLags[hosts[i]]
				x.DataPoints = append(x.DataPoints, getDataPoint(v, t))
				replicationLags[hosts[i]] = x
			}
		}
	}

	*legends = hosts
	return timeSeriesData, replicationLags
}

func getSystemMetricsTimeSeriesDoc(systemMetricsList []SystemMetricsDoc) (map[string]TimeSeriesDoc, map[string]DiskStats) {
	var timeSeriesData = map[string]TimeSeriesDoc{}
	var diskStats = map[string]DiskStats{}
	var pstat = SystemMetricsDoc{}

	for _, legend := range systemMetricsChartsLegends {
		timeSeriesData[legend] = TimeSeriesDoc{legend, [][]float64{}}
	}
	for i, stat := range systemMetricsList {
		if i > 0 {
			t := float64(stat.Start.UnixNano() / (1000 * 1000))
			for k, disk := range stat.Disks {
				u := 100 * float64(disk.IOTimeMS-pstat.Disks[k].IOTimeMS) / 1000 // / 1000 ms * 100 %
				iops := float64(disk.Reads+disk.Writes-(pstat.Disks[k].Reads+pstat.Disks[k].Writes)) / float64(stat.Start.Sub(pstat.Start).Seconds())
				qlen := float64(disk.IOInProgress)
				readTimeMS := float64(disk.ReadTimeMS - pstat.Disks[k].ReadTimeMS)
				writeTimeMS := float64(disk.WriteTimeMS - pstat.Disks[k].WriteTimeMS)
				ioQueuedMS := float64(disk.IOQueuedMS - pstat.Disks[k].IOQueuedMS)
				x := diskStats[k]
				x.Utilization.DataPoints = append(x.Utilization.DataPoints, getDataPoint(u, t))
				x.IOPS.DataPoints = append(x.IOPS.DataPoints, getDataPoint(iops, t))
				x.IOInProgress.DataPoints = append(x.IOInProgress.DataPoints, getDataPoint(qlen, t))
				x.ReadTimeMS.DataPoints = append(x.ReadTimeMS.DataPoints, getDataPoint(readTimeMS, t))
				x.WriteTimeMS.DataPoints = append(x.WriteTimeMS.DataPoints, getDataPoint(writeTimeMS, t))
				x.IOQueuedMS.DataPoints = append(x.IOQueuedMS.DataPoints, getDataPoint(ioQueuedMS, t))
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

func getServerStatusTimeSeriesDoc(serverStatusList []ServerStatusDoc) map[string]TimeSeriesDoc {
	var timeSeriesData = map[string]TimeSeriesDoc{}
	pstat := ServerStatusDoc{}
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

			x = timeSeriesData["conns_active"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Connections.Active), t))
			timeSeriesData["conns_active"] = x

			x = timeSeriesData["conns_available"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Connections.Available), t))
			timeSeriesData["conns_available"] = x

			x = timeSeriesData["conns_current"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Connections.Current), t))
			timeSeriesData["conns_current"] = x

			x = timeSeriesData["q_active_read"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.GlobalLock.ActiveClients.Readers), t))
			timeSeriesData["q_active_read"] = x

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

			x = timeSeriesData["q_active_write"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.GlobalLock.ActiveClients.Writers), t))
			timeSeriesData["q_active_write"] = x

			x = timeSeriesData["q_queued_read"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.GlobalLock.CurrentQueue.Readers), t))
			timeSeriesData["q_queued_read"] = x

			x = timeSeriesData["q_queued_write"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.GlobalLock.CurrentQueue.Writers), t))
			timeSeriesData["q_queued_write"] = x

			x = timeSeriesData["tcmalloc_current_allocated_bytes"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.TCMalloc.Generic.CurrentAllocatedBytes), t))
			timeSeriesData["tcmalloc_current_allocated_bytes"] = x

			x = timeSeriesData["tcmalloc_heap_size"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.TCMalloc.Generic.HeapSize), t))
			timeSeriesData["tcmalloc_heap_size"] = x

			x = timeSeriesData["tcmalloc_pageheap_free_bytes"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.TCMalloc.TCMalloc.PageheapFreeBytes), t))
			timeSeriesData["tcmalloc_pageheap_free_bytes"] = x

			x = timeSeriesData["tcmalloc_pageheap_unmapped_bytes"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.TCMalloc.TCMalloc.PageheapUnmappedBytes), t))
			timeSeriesData["tcmalloc_pageheap_unmapped_bytes"] = x

			x = timeSeriesData["tcmalloc_max_total_thread_cache_bytes"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.TCMalloc.TCMalloc.MaxTotalThreadCacheBytes), t))
			timeSeriesData["tcmalloc_max_total_thread_cache_bytes"] = x

			x = timeSeriesData["tcmalloc_current_total_thread_cache_bytes"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.TCMalloc.TCMalloc.CurrentTotalThreadCacheBytes), t))
			timeSeriesData["tcmalloc_current_total_thread_cache_bytes"] = x

			x = timeSeriesData["tcmalloc_total_free_bytes"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.TCMalloc.TCMalloc.TotalFreeBytes), t))
			timeSeriesData["tcmalloc_total_free_bytes"] = x

			x = timeSeriesData["tcmalloc_central_cache_free_bytes"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.TCMalloc.TCMalloc.CentralCacheFreeBytes), t))
			timeSeriesData["tcmalloc_central_cache_free_bytes"] = x

			x = timeSeriesData["tcmalloc_transfer_cache_free_bytes"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.TCMalloc.TCMalloc.TransferCacheFreeBytes), t))
			timeSeriesData["tcmalloc_transfer_cache_free_bytes"] = x

			x = timeSeriesData["tcmalloc_thread_cache_free_bytes"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.TCMalloc.TCMalloc.ThreadCacheFreeBytes), t))
			timeSeriesData["tcmalloc_thread_cache_free_bytes"] = x

			x = timeSeriesData["tcmalloc_aggressive_memory_decommit"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.TCMalloc.TCMalloc.AggressiveMemoryDecommit), t))
			timeSeriesData["tcmalloc_aggressive_memory_decommit"] = x

			x = timeSeriesData["tcmalloc_pageheap_committed_bytes"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.TCMalloc.TCMalloc.PageheapComittedBytes), t))
			timeSeriesData["tcmalloc_pageheap_committed_bytes"] = x

			x = timeSeriesData["tcmalloc_release_rate"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.TCMalloc.TCMalloc.ReleaseRate), t))
			timeSeriesData["tcmalloc_release_rate"] = x

			if i > 0 {
				seconds := math.Round(stat.LocalTime.Sub(pstat.LocalTime).Seconds())
				if seconds < 1 {
					seconds = 1
				}

				x = timeSeriesData["mem_page_faults"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.ExtraInfo.PageFaults-pstat.ExtraInfo.PageFaults)/seconds, t))
				timeSeriesData["mem_page_faults"] = x

				x = timeSeriesData["conns_created/s"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Connections.TotalCreated-pstat.Connections.TotalCreated)/seconds, t))
				timeSeriesData["conns_created/s"] = x

				x = timeSeriesData["net_in"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Network.BytesIn-pstat.Network.BytesIn)/mb/seconds, t))
				timeSeriesData["net_in"] = x

				x = timeSeriesData["net_out"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Network.BytesOut-pstat.Network.BytesOut)/mb/seconds, t))
				timeSeriesData["net_out"] = x

				x = timeSeriesData["net_requests"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Network.NumRequests-pstat.Network.NumRequests)/seconds, t))
				timeSeriesData["net_requests"] = x

				x = timeSeriesData["net_physical_in"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Network.PhysicalBytesIn-pstat.Network.PhysicalBytesIn)/mb/seconds, t))
				timeSeriesData["net_physicalsical_in"] = x

				x = timeSeriesData["net_physical_out"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Network.PhysicalBytesOut-pstat.Network.PhysicalBytesOut)/mb/seconds, t))
				timeSeriesData["net_physical_out"] = x

				x = timeSeriesData["ops_query"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.OpCounters.Query-pstat.OpCounters.Query)/seconds, t))
				timeSeriesData["ops_query"] = x

				x = timeSeriesData["ops_insert"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.OpCounters.Insert-pstat.OpCounters.Insert)/seconds, t))
				timeSeriesData["ops_insert"] = x

				x = timeSeriesData["ops_update"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.OpCounters.Update-pstat.OpCounters.Update)/seconds, t))
				timeSeriesData["ops_update"] = x

				x = timeSeriesData["ops_delete"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.OpCounters.Delete-pstat.OpCounters.Delete)/seconds, t))
				timeSeriesData["ops_delete"] = x

				x = timeSeriesData["ops_getmore"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.OpCounters.Getmore-pstat.OpCounters.Getmore)/seconds, t))
				timeSeriesData["ops_getmore"] = x

				x = timeSeriesData["ops_command"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.OpCounters.Command-pstat.OpCounters.Command)/seconds, t))
				timeSeriesData["ops_command"] = x

				x = timeSeriesData["scan_keys"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Metrics.QueryExecutor.Scanned-pstat.Metrics.QueryExecutor.Scanned)/seconds, t))
				timeSeriesData["scan_keys"] = x

				x = timeSeriesData["scan_objects"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Metrics.QueryExecutor.ScannedObjects-pstat.Metrics.QueryExecutor.ScannedObjects)/seconds, t))
				timeSeriesData["scan_objects"] = x

				x = timeSeriesData["scan_sort"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.Metrics.Operation.ScanAndOrder-pstat.Metrics.Operation.ScanAndOrder)/seconds, t))
				timeSeriesData["scan_sort"] = x

				x = timeSeriesData["tcmalloc_spinlock_total_delay_ns"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.TCMalloc.TCMalloc.SpinLockTotalDelayNanos-pstat.TCMalloc.TCMalloc.SpinLockTotalDelayNanos)/seconds, t))
				timeSeriesData["tcmalloc_spinlock_total_delay_ns"] = x

				x = timeSeriesData["tcmalloc_pageheap_commit_count"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.TCMalloc.TCMalloc.PageheapCommitCount-pstat.TCMalloc.TCMalloc.PageheapCommitCount)/seconds, t))
				timeSeriesData["tcmalloc_pageheap_commit_count"] = x

				x = timeSeriesData["tcmalloc_pageheap_total_commit_bytes"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.TCMalloc.TCMalloc.PageheapTotalCommitBytes-pstat.TCMalloc.TCMalloc.PageheapTotalCommitBytes)/seconds, t))
				timeSeriesData["tcmalloc_pageheap_total_commit_bytes"] = x

				x = timeSeriesData["tcmalloc_pageheap_decommit_count"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.TCMalloc.TCMalloc.PageheapDecommitCount-pstat.TCMalloc.TCMalloc.PageheapDecommitCount)/seconds, t))
				timeSeriesData["tcmalloc_pageheap_decommit_count"] = x

				x = timeSeriesData["tcmalloc_pageheap_total_decommit_bytes"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.TCMalloc.TCMalloc.PageheapTotalDecommitBytes-pstat.TCMalloc.TCMalloc.PageheapTotalDecommitBytes)/seconds, t))
				timeSeriesData["tcmalloc_pageheap_total_decommit_bytes"] = x

				x = timeSeriesData["tcmalloc_pageheap_reserve_count"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.TCMalloc.TCMalloc.PageheapReserveCount-pstat.TCMalloc.TCMalloc.PageheapReserveCount)/seconds, t))
				timeSeriesData["tcmalloc_pageheap_reserve_count"] = x

				x = timeSeriesData["tcmalloc_pageheap_total_reserve_bytes"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.TCMalloc.TCMalloc.PageheapTotalReserveBytes-pstat.TCMalloc.TCMalloc.PageheapTotalReserveBytes)/seconds, t))
				timeSeriesData["tcmalloc_pageheap_total_reserve_bytes"] = x

				x = timeSeriesData["tcmalloc_pageheap_scavenge_count"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.TCMalloc.TCMalloc.PageheapScavengeCount-pstat.TCMalloc.TCMalloc.PageheapScavengeCount)/seconds, t))
				timeSeriesData["tcmalloc_pageheap_scavenge_count"] = x
			} // if i > 0
		} // if stat.Uptime > pstat.Uptime

		pstat = stat
	}
	return timeSeriesData
}

func getWiredTigerTimeSeriesDoc(serverStatusList []ServerStatusDoc) map[string]TimeSeriesDoc {
	var timeSeriesData = map[string]TimeSeriesDoc{}
	pstat := ServerStatusDoc{}
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

			x = timeSeriesData["wt_dhandles_active"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.DataHandle.Active), t))
			timeSeriesData["wt_dhandles_active"] = x

			x = timeSeriesData["ticket_avail_read"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.ConcurrentTransactions.Read.Available), t))
			timeSeriesData["ticket_avail_read"] = x

			x = timeSeriesData["ticket_avail_write"]
			x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.ConcurrentTransactions.Write.Available), t))
			timeSeriesData["ticket_avail_write"] = x

			if i > 0 {
				seconds := math.Round(stat.LocalTime.Sub(pstat.LocalTime).Seconds())
				if seconds < 1 {
					seconds = 1
				}
				x = timeSeriesData["wt_blkmgr_read"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.BlockManager.BytesRead-pstat.WiredTiger.BlockManager.BytesRead)/mb/seconds, t))
				timeSeriesData["wt_blkmgr_read"] = x

				x = timeSeriesData["wt_blkmgr_written"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.BlockManager.BytesWritten-pstat.WiredTiger.BlockManager.BytesWritten)/mb/seconds, t))
				timeSeriesData["wt_blkmgr_written"] = x

				x = timeSeriesData["wt_blkmgr_written_checkpoint"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.BlockManager.BytesWrittenCheckPoint-pstat.WiredTiger.BlockManager.BytesWrittenCheckPoint)/mb/seconds, t))
				timeSeriesData["wt_blkmgr_written_checkpoint"] = x

				x = timeSeriesData["wt_modified_evicted"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.Cache.ModifiedPagesEvicted-pstat.WiredTiger.Cache.ModifiedPagesEvicted)/seconds, t))
				timeSeriesData["wt_modified_evicted"] = x

				x = timeSeriesData["wt_unmodified_evicted"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.Cache.UnmodifiedPagesEvicted-pstat.WiredTiger.Cache.UnmodifiedPagesEvicted)/seconds, t))
				timeSeriesData["wt_unmodified_evicted"] = x

				x = timeSeriesData["wt_cache_read_in"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.Cache.BytesReadIntoCache-pstat.WiredTiger.Cache.BytesReadIntoCache)/mb/seconds, t))
				timeSeriesData["wt_cache_read_in"] = x

				x = timeSeriesData["wt_cache_written_from"]
				x.DataPoints = append(x.DataPoints, getDataPoint(float64(stat.WiredTiger.Cache.BytesWrittenFromCache-pstat.WiredTiger.Cache.BytesWrittenFromCache)/mb/seconds, t))
				timeSeriesData["wt_cache_written_from"] = x
			} // if i > 0
		} // if stat.Uptime > pstat.Uptime

		pstat = stat
	}
	return timeSeriesData
}
