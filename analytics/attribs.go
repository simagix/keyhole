// Copyright 2020 Kuei-chun Chen. All rights reserved.

package analytics

import (
	"math"
	"strings"
	"time"

	"github.com/simagix/keyhole/ftdc"
)

// Attribs stores attribs map
type Attribs struct {
	attribsMap *map[string][]uint64
}

// NewAttribs returns Attribs structure
func NewAttribs(attribsMap *map[string][]uint64) *Attribs {
	return &Attribs{attribsMap: attribsMap}
}

// GetServerStatusDataPoints returns server status
func (attr *Attribs) GetServerStatusDataPoints(i int) ServerStatusDoc {
	ss := ServerStatusDoc{}
	ss.LocalTime = time.Unix(0, int64(time.Millisecond)*int64(attr.get("serverStatus/localTime", i)))
	ss.Mem.Resident = attr.get("serverStatus/mem/resident", i)
	ss.Mem.Virtual = attr.get("serverStatus/mem/virtual", i)
	ss.Network.BytesIn = attr.get("serverStatus/network/bytesIn", i)
	ss.Network.BytesOut = attr.get("serverStatus/network/bytesOut", i)
	ss.Network.NumRequests = attr.get("serverStatus/network/numRequests", i)
	ss.Network.PhysicalBytesIn = attr.get("serverStatus/network/physicalBytesIn", i)
	ss.Network.PhysicalBytesOut = attr.get("serverStatus/network/physicalBytesOut", i)
	ss.Connections.Current = attr.get("serverStatus/connections/current", i)
	ss.Connections.TotalCreated = attr.get("serverStatus/connections/totalCreated", i)
	ss.Connections.Available = attr.get("serverStatus/connections/available", i)
	ss.Connections.Active = attr.get("serverStatus/connections/active", i)
	ss.ExtraInfo.PageFaults = attr.get("serverStatus/extra_info/page_faults", i)
	ss.GlobalLock.ActiveClients.Readers = attr.get("serverStatus/globalLock/activeClients/readers", i)
	ss.GlobalLock.ActiveClients.Writers = attr.get("serverStatus/globalLock/activeClients/writers", i)
	ss.GlobalLock.CurrentQueue.Readers = attr.get("serverStatus/globalLock/currentQueue/readers", i)
	ss.GlobalLock.CurrentQueue.Writers = attr.get("serverStatus/globalLock/currentQueue/writers", i)
	ss.Metrics.QueryExecutor.Scanned = attr.get("serverStatus/metrics/queryExecutor/scanned", i)
	ss.Metrics.QueryExecutor.ScannedObjects = attr.get("serverStatus/metrics/queryExecutor/scannedObjects", i)
	ss.Metrics.Operation.ScanAndOrder = attr.get("serverStatus/metrics/operation/scanAndOrder", i)
	ss.OpLatencies.Commands.Latency = attr.get("serverStatus/opLatencies/commands/latency", i)
	ss.OpLatencies.Commands.Ops = attr.get("serverStatus/opLatencies/commands/ops", i)
	ss.OpLatencies.Reads.Latency = attr.get("serverStatus/opLatencies/reads/latency", i)
	ss.OpLatencies.Reads.Ops = attr.get("serverStatus/opLatencies/reads/ops", i)
	ss.OpLatencies.Writes.Latency = attr.get("serverStatus/opLatencies/writes/latency", i)
	ss.OpLatencies.Writes.Ops = attr.get("serverStatus/opLatencies/writes/ops", i)
	ss.OpCounters.Command = attr.get("serverStatus/opcounters/command", i)
	ss.OpCounters.Delete = attr.get("serverStatus/opcounters/delete", i)
	ss.OpCounters.Getmore = attr.get("serverStatus/opcounters/getmore", i)
	ss.OpCounters.Insert = attr.get("serverStatus/opcounters/insert", i)
	ss.OpCounters.Query = attr.get("serverStatus/opcounters/query", i)
	ss.OpCounters.Update = attr.get("serverStatus/opcounters/update", i)
	ss.Uptime = attr.get("serverStatus/uptime", i)

	ss.WiredTiger.BlockManager.BytesRead = attr.get("serverStatus/wiredTiger/block-manager/bytes read", i)
	ss.WiredTiger.BlockManager.BytesWritten = attr.get("serverStatus/wiredTiger/block-manager/bytes written", i)
	ss.WiredTiger.BlockManager.BytesWrittenCheckPoint = attr.get("serverStatus/wiredTiger/block-manager/bytes written for checkpoint", i)
	ss.WiredTiger.Cache.CurrentlyInCache = attr.get("serverStatus/wiredTiger/cache/bytes currently in the cache", i)
	ss.WiredTiger.Cache.MaxBytesConfigured = attr.get("serverStatus/wiredTiger/cache/maximum bytes configured", i)
	ss.WiredTiger.Cache.ModifiedPagesEvicted = attr.get("serverStatus/wiredTiger/cache/modified pages evicted", i)
	ss.WiredTiger.Cache.BytesReadIntoCache = attr.get("serverStatus/wiredTiger/cache/bytes read into cache", i)
	ss.WiredTiger.Cache.BytesWrittenFromCache = attr.get("serverStatus/wiredTiger/cache/bytes written from cache", i)
	ss.WiredTiger.Cache.TrackedDirtyBytes = attr.get("serverStatus/wiredTiger/cache/tracked dirty bytes in the cache", i)
	ss.WiredTiger.Cache.UnmodifiedPagesEvicted = attr.get("serverStatus/wiredTiger/cache/unmodified pages evicted", i)
	ss.WiredTiger.DataHandle.Active = attr.get("serverStatus/wiredTiger/data-handle/connection data handles currently active", i)
	ss.WiredTiger.ConcurrentTransactions.Read.Available = attr.get("serverStatus/wiredTiger/concurrentTransactions/read/available", i)
	ss.WiredTiger.ConcurrentTransactions.Write.Available = attr.get("serverStatus/wiredTiger/concurrentTransactions/write/available", i)

	ss.TCMalloc.Generic.CurrentAllocatedBytes = attr.get("serverStatus/tcmalloc/generic/current_allocated_bytes", i)
	ss.TCMalloc.Generic.HeapSize = attr.get("serverStatus/tcmalloc/generic/heap_size", i)
	ss.TCMalloc.TCMalloc.PageheapFreeBytes = attr.get("serverStatus/tcmalloc/tcmalloc/pageheap_free_bytes", i)
	ss.TCMalloc.TCMalloc.PageheapUnmappedBytes = attr.get("serverStatus/tcmalloc/tcmalloc/pageheap_unmapped_bytes", i)
	ss.TCMalloc.TCMalloc.MaxTotalThreadCacheBytes = attr.get("serverStatus/tcmalloc/tcmalloc/max_total_thread_cache_bytes", i)
	ss.TCMalloc.TCMalloc.CurrentTotalThreadCacheBytes = attr.get("serverStatus/tcmalloc/tcmalloc/current_total_thread_cache_bytes", i)
	ss.TCMalloc.TCMalloc.TotalFreeBytes = attr.get("serverStatus/tcmalloc/tcmalloc/total_free_bytes", i)
	ss.TCMalloc.TCMalloc.CentralCacheFreeBytes = attr.get("serverStatus/tcmalloc/tcmalloc/central_cache_free_bytes", i)
	ss.TCMalloc.TCMalloc.TransferCacheFreeBytes = attr.get("serverStatus/tcmalloc/tcmalloc/transfer_cache_free_bytes", i)
	ss.TCMalloc.TCMalloc.ThreadCacheFreeBytes = attr.get("serverStatus/tcmalloc/tcmalloc/thread_cache_free_bytes", i)
	ss.TCMalloc.TCMalloc.AggressiveMemoryDecommit = attr.get("serverStatus/tcmalloc/tcmalloc/aggressive_memory_decommit", i)
	ss.TCMalloc.TCMalloc.PageheapComittedBytes = attr.get("serverStatus/tcmalloc/tcmalloc/pageheap_committed_bytes", i)
	ss.TCMalloc.TCMalloc.PageheapScavengeCount = attr.get("serverStatus/tcmalloc/tcmalloc/pageheap_scavenge_count", i)
	ss.TCMalloc.TCMalloc.PageheapCommitCount = attr.get("serverStatus/tcmalloc/tcmalloc/pageheap_commit_count", i)
	ss.TCMalloc.TCMalloc.PageheapTotalCommitBytes = attr.get("serverStatus/tcmalloc/tcmalloc/pageheap_total_commit_bytes", i)
	ss.TCMalloc.TCMalloc.PageheapDecommitCount = attr.get("serverStatus/tcmalloc/tcmalloc/pageheap_decommit_count", i)
	ss.TCMalloc.TCMalloc.PageheapTotalDecommitBytes = attr.get("serverStatus/tcmalloc/tcmalloc/pageheap_total_decommit_bytes", i)
	ss.TCMalloc.TCMalloc.PageheapReserveCount = attr.get("serverStatus/tcmalloc/tcmalloc/pageheap_reserve_count", i)
	ss.TCMalloc.TCMalloc.PageheapTotalReserveBytes = attr.get("serverStatus/tcmalloc/tcmalloc/pageheap_total_reserve_bytes", i)
	ss.TCMalloc.TCMalloc.SpinLockTotalDelayNanos = attr.get("serverStatus/tcmalloc/tcmalloc/spinlock_total_delay_ns", i)
	ss.TCMalloc.TCMalloc.ReleaseRate = attr.get("serverStatus/tcmalloc/tcmalloc/release_rate", i)

	return ss
}

// GetSystemMetricsDataPoints returns system metrics
func (attr *Attribs) GetSystemMetricsDataPoints(i int) SystemMetricsDoc {
	attribs := *attr.attribsMap
	sm := SystemMetricsDoc{Disks: map[string]DiskMetrics{}}
	sm.Start = time.Unix(0, int64(time.Millisecond)*int64(attr.get("serverStatus/localTime", i)))
	sm.CPU.IdleMS = attr.get("systemMetrics/cpu/idle_ms", i)
	sm.CPU.UserMS = attr.get("systemMetrics/cpu/user_ms", i)
	sm.CPU.IOWaitMS = attr.get("systemMetrics/cpu/iowait_ms", i)
	sm.CPU.NiceMS = attr.get("systemMetrics/cpu/nice_ms", i)
	sm.CPU.SoftirqMS = attr.get("systemMetrics/cpu/softirq_ms", i)
	sm.CPU.StealMS = attr.get("systemMetrics/cpu/steal_ms", i)
	sm.CPU.SystemMS = attr.get("systemMetrics/cpu/system_ms", i)
	for key := range attribs {
		if strings.Index(key, "systemMetrics/disks/") != 0 {
			continue
		}
		tokens := strings.Split(key, ftdc.PathSeparator)
		disk := tokens[2]
		stats := tokens[3]
		if _, ok := sm.Disks[disk]; !ok {
			sm.Disks[disk] = DiskMetrics{}
		}
		m := sm.Disks[disk]
		switch stats {
		case "read_time_ms":
			m.ReadTimeMS = attr.get(key, i)
		case "write_time_ms":
			m.WriteTimeMS = attr.get(key, i)
		case "io_queued_ms":
			m.IOQueuedMS = attr.get(key, i)
		case "io_time_ms":
			m.IOTimeMS = attr.get(key, i)
		case "reads":
			m.Reads = attr.get(key, i)
		case "writes":
			m.Writes = attr.get(key, i)
		case "io_in_progress":
			m.IOInProgress = attr.get(key, i)
		}
		sm.Disks[disk] = m
	}
	return sm
}

func (attr *Attribs) get(key string, i int) uint64 {
	arr := (*attr.attribsMap)[key]
	if i < len(arr) && !math.IsNaN(float64(arr[i])) {
		return arr[i]
	}
	return 0
}
