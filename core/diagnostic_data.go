// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"strings"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/simagix/keyhole/ftdc"
)

// decodeFTDC decodes FTDC data.
func getDiagnosticData(data []byte, span int) (DiagnosticData, error) {
	var err error
	var diagData = DiagnosticData{}

	if span >= 300 {
		return unmarshalFirstBsonDoc(data), err
	}

	f := ftdc.NewFTDC()
	if err = f.Decode(data); err != nil {
		return diagData, err
	}

	for i := uint32(0); i < f.NumDeltas; i += uint32(span) {
		ss, sm := getDataPoints(f.DataPointsMap, i)
		diagData.ServerStatusList = append(diagData.ServerStatusList, ss)
		diagData.SystemMetricsList = append(diagData.SystemMetricsList, sm)
	}

	// replSetGetStatus
	var doc DiagnosticDoc
	bson.Unmarshal(data, &doc) // first document
	repl := doc.ReplSetGetStatus
	for _, mb := range repl.Members {
		mb.Optime = GetOptime(mb.Optime)
	}
	diagData.ReplSetStatusList = append(diagData.ReplSetStatusList, repl)
	return diagData, nil
}

func getDataPoints(attribsMap map[string][]int64, i uint32) (ServerStatusDoc, SystemMetricsDoc) {
	ss := ServerStatusDoc{}
	ss.LocalTime = time.Unix(0, int64(time.Millisecond)*attribsMap["serverStatus/localTime"][i])
	ss.Mem.Resident = attribsMap["serverStatus/mem/resident"][i]
	ss.Mem.Virtual = attribsMap["serverStatus/mem/virtual"][i]
	ss.Connections.Current = attribsMap["serverStatus/connections/current"][i]
	ss.Connections.TotalCreated = attribsMap["serverStatus/connections/totalCreated"][i]
	ss.ExtraInfo.PageFaults = attribsMap["serverStatus/extra_info/page_faults"][i]
	ss.GlobalLock.ActiveClients.Readers = attribsMap["serverStatus/globalLock/activeClients/readers"][i]
	ss.GlobalLock.ActiveClients.Writers = attribsMap["serverStatus/globalLock/activeClients/writers"][i]
	ss.GlobalLock.CurrentQueue.Readers = attribsMap["serverStatus/globalLock/currentQueue/readers"][i]
	ss.GlobalLock.CurrentQueue.Writers = attribsMap["serverStatus/globalLock/currentQueue/writers"][i]
	ss.Metrics.QueryExecutor.Scanned = attribsMap["serverStatus/metrics/queryExecutor/scanned"][i]
	ss.Metrics.QueryExecutor.ScannedObjects = attribsMap["serverStatus/metrics/queryExecutor/scannedObjects"][i]
	ss.Metrics.Operation.ScanAndOrder = attribsMap["serverStatus/metrics/operation/scanAndOrder"][i]
	if len(attribsMap["serverStatus/opLatencies/commands/latency"]) > 1 { // 3.2 didn't have opLatencies
		ss.OpLatencies.Commands.Latency = attribsMap["serverStatus/opLatencies/commands/latency"][i]
		ss.OpLatencies.Commands.Ops = attribsMap["serverStatus/opLatencies/commands/ops"][i]
		ss.OpLatencies.Reads.Latency = attribsMap["serverStatus/opLatencies/reads/latency"][i]
		ss.OpLatencies.Reads.Ops = attribsMap["serverStatus/opLatencies/reads/ops"][i]
		ss.OpLatencies.Writes.Latency = attribsMap["serverStatus/opLatencies/writes/latency"][i]
		ss.OpLatencies.Writes.Ops = attribsMap["serverStatus/opLatencies/writes/ops"][i]
	}
	ss.OpCounters.Command = attribsMap["serverStatus/opcounters/command"][i]
	ss.OpCounters.Delete = attribsMap["serverStatus/opcounters/delete"][i]
	ss.OpCounters.Getmore = attribsMap["serverStatus/opcounters/getmore"][i]
	ss.OpCounters.Insert = attribsMap["serverStatus/opcounters/insert"][i]
	ss.OpCounters.Query = attribsMap["serverStatus/opcounters/query"][i]
	ss.OpCounters.Update = attribsMap["serverStatus/opcounters/update"][i]
	ss.Uptime = attribsMap["serverStatus/uptime"][i]
	ss.WiredTiger.Cache.CurrentlyInCache = attribsMap["serverStatus/wiredTiger/cache/bytes currently in the cache"][i]
	ss.WiredTiger.Cache.MaxBytesConfigured = attribsMap["serverStatus/wiredTiger/cache/maximum bytes configured"][i]
	ss.WiredTiger.Cache.ModifiedPagesEvicted = attribsMap["serverStatus/wiredTiger/cache/modified pages evicted"][i]
	ss.WiredTiger.Cache.PagesReadIntoCache = attribsMap["serverStatus/wiredTiger/cache/pages read into cache"][i]
	ss.WiredTiger.Cache.PagesWrittenFromCache = attribsMap["serverStatus/wiredTiger/cache/pages written from cache"][i]
	ss.WiredTiger.Cache.TrackedDirtyBytes = attribsMap["serverStatus/wiredTiger/cache/tracked dirty bytes in the cache"][i]
	ss.WiredTiger.Cache.UnmodifiedPagesEvicted = attribsMap["serverStatus/wiredTiger/cache/unmodified pages evicted"][i]
	ss.WiredTiger.ConcurrentTransactions.Read.Available = attribsMap["serverStatus/wiredTiger/concurrentTransactions/read/available"][i]
	ss.WiredTiger.ConcurrentTransactions.Write.Available = attribsMap["serverStatus/wiredTiger/concurrentTransactions/write/available"][i]

	// systemMetrics
	sm := SystemMetricsDoc{}
	sm.Start = ss.LocalTime
	sm.CPU.IdleMS = attribsMap["systemMetrics/cpu/idle_ms"][i]
	sm.CPU.UserMS = attribsMap["systemMetrics/cpu/user_ms"][i]
	sm.CPU.IOWaitMS = attribsMap["systemMetrics/cpu/iowait_ms"][i]
	sm.CPU.NiceMS = attribsMap["systemMetrics/cpu/nice_ms"][i]
	sm.CPU.SoftirqMS = attribsMap["systemMetrics/cpu/softirq_ms"][i]
	sm.CPU.StealMS = attribsMap["systemMetrics/cpu/steal_ms"][i]
	sm.CPU.SystemMS = attribsMap["systemMetrics/cpu/system_ms"][i]

	diskMap := map[string]DiskMetrics{}
	for key, _ := range attribsMap {
		if strings.Index(key, "systemMetrics/disks/") != 0 {
			continue
		}
		tokens := strings.Split(key, ftdc.PathSeparator)
		if _, ok := diskMap[tokens[2]]; !ok {
			diskMap[tokens[2]] = DiskMetrics{}
		}
		m := diskMap[tokens[2]]
		switch tokens[3] {
		case "read_time_ms":
			m.ReadTimeMS = attribsMap[key][i]
		case "write_time_ms":
			m.WriteTimeMS = attribsMap[key][i]
		case "io_queued_ms":
			m.IOQueuedMS = attribsMap[key][i]
		case "io_time_ms":
			m.IOTimeMS = attribsMap[key][i]
		case "reads":
			m.Reads = attribsMap[key][i]
		case "writes":
			m.Writes = attribsMap[key][i]
		}
		diskMap[tokens[2]] = m
	}
	sm.Disks = diskMap
	return ss, sm
}
