// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"bytes"
	"errors"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/globalsign/mgo/bson"
)

const pathSeparator = "/" // path separator

// decodeFTDC decodes FTDC data.
func getDiagnosticData(data []byte, span int) (DiagnosticData, error) {
	var err error

	if span >= 300 {
		return unmarshalFirstBsonDoc(data), err
	}

	var diagData = DiagnosticData{}
	r := bytes.NewReader(data)
	length := GetUint32(r) // first bson document length
	r.Seek(int64(length), io.SeekStart)
	nAttribs := GetUint32(r) // 4 bytes # of keys
	nDeltas := GetUint32(r)  // 4 bytes # of deltas
	ptr, _ := r.Seek(0, io.SeekCurrent)
	deltas := data[ptr:]

	// use DOM (bson.D) to ensure orders
	var attribsMap = map[string][]int64{}
	var attribsList = []string{}
	var docElem = bson.D{}
	bson.Unmarshal(data, &docElem) // first document
	traverseDocElem(&attribsList, &attribsMap, docElem, "")
	if len(attribsMap) != int(nAttribs) || len(attribsList) != int(nAttribs) {
		return diagData, errors.New("inconsistent diagnostic data")
	}

	// deltas
	// d where d > 0, return d
	// 0d -> there are d number of zeros
	var delta int64
	r = bytes.NewReader(deltas)
	var zerosLeft uint64
	for i := uint32(0); i < nAttribs; i++ {
		list := attribsMap[attribsList[i]]
		v := list[0]
		for j := uint32(0); j < nDeltas; j++ {
			if zerosLeft != 0 {
				delta = 0
				zerosLeft--
			} else {
				delta = int64(Uvarint(r))
				if delta == 0 {
					zerosLeft = Uvarint(r)
				}
			}
			v += delta
			list = append(list, v)
		}
		attribsMap[attribsList[i]] = list
	}

	for i := uint32(0); i < nDeltas; i += uint32(span) {
		ss, sm := getDataPoints(attribsMap, i, attribsList)
		diagData.ServerStatusList = append(diagData.ServerStatusList, ss)
		diagData.SystemMetricsList = append(diagData.SystemMetricsList, sm)
	}

	// replSetGetStatus
	var doc DiagnosticDoc
	bson.Unmarshal(data, &doc) // first document
	diagData.ReplSetStatusList = append(diagData.ReplSetStatusList, doc.ReplSetGetStatus)
	return diagData, nil
}

func getDataPoints(attribsMap map[string][]int64, i uint32, attribsList []string) (ServerStatusDoc, SystemMetricsDoc) {
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
	ss.OpLatencies.Commands.Latency = attribsMap["serverStatus/opLatencies/commands/latency"][i]
	ss.OpLatencies.Commands.Ops = attribsMap["serverStatus/opLatencies/commands/ops"][i]
	ss.OpLatencies.Reads.Latency = attribsMap["serverStatus/opLatencies/reads/latency"][i]
	ss.OpLatencies.Reads.Ops = attribsMap["serverStatus/opLatencies/reads/ops"][i]
	ss.OpLatencies.Writes.Latency = attribsMap["serverStatus/opLatencies/writes/latency"][i]
	ss.OpLatencies.Writes.Ops = attribsMap["serverStatus/opLatencies/writes/ops"][i]
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
	for _, str := range attribsList {
		if strings.Index(str, "systemMetrics/disks/") != 0 {
			continue
		}
		tokens := strings.Split(str, pathSeparator)
		if _, ok := diskMap[tokens[2]]; !ok {
			diskMap[tokens[2]] = DiskMetrics{}
		}
		m := diskMap[tokens[2]]
		switch tokens[3] {
		case "read_time_ms":
			m.ReadTimeMS = attribsMap[str][i]
		case "write_time_ms":
			m.WriteTimeMS = attribsMap[str][i]
		case "io_queued_ms":
			m.IOQueuedMS = attribsMap[str][i]
		case "io_time_ms":
			m.IOTimeMS = attribsMap[str][i]
		case "reads":
			m.Reads = attribsMap[str][i]
		case "writes":
			m.Writes = attribsMap[str][i]
		}
		diskMap[tokens[2]] = m
	}
	sm.Disks = diskMap
	return ss, sm
}

func traverseDocElem(attribsList *[]string, attribsMap *map[string][]int64, docElem interface{}, parentPath string) {
	switch value := docElem.(type) {
	case []interface{}:
		for i, v := range value {
			fld := parentPath + pathSeparator + strconv.Itoa(i)
			traverseDocElem(attribsList, attribsMap, v, fld)
		}
	case bool:
		v := int64(0)
		if value == true {
			v = 1
		}
		x := []int64{v}
		(*attribsMap)[parentPath] = x
		(*attribsList) = append((*attribsList), parentPath)
	case bson.D:
		elem := docElem.(bson.D)
		for _, elem := range elem {
			name := elem.Name
			if parentPath != "" {
				name = parentPath + pathSeparator + name
			}
			traverseDocElem(attribsList, attribsMap, elem.Value, name)
		}
	case bson.MongoTimestamp:
		tKey := parentPath + "/t"
		iKey := parentPath + "/i"
		(*attribsMap)[tKey] = []int64{int64(value)}
		(*attribsList) = append((*attribsList), tKey)
		(*attribsMap)[iKey] = []int64{int64(value)}
		(*attribsList) = append((*attribsList), iKey)
	case bson.ObjectId, string: // ignore it
		// (*attribsMap)[parentPath] = []int{}
	case float64:
		(*attribsMap)[parentPath] = []int64{int64(value)}
		(*attribsList) = append((*attribsList), parentPath)
	case int:
		(*attribsMap)[parentPath] = []int64{int64(value)}
		(*attribsList) = append((*attribsList), parentPath)
	case int64:
		(*attribsMap)[parentPath] = []int64{value}
		(*attribsList) = append((*attribsList), parentPath)
	case time.Time:
		(*attribsMap)[parentPath] = []int64{time.Time(value).UnixNano() / 1000000}
		(*attribsList) = append((*attribsList), parentPath)
	default:
		log.Fatalf("%s %T\n", parentPath, value)
	}
}

// Uvarint -
func Uvarint(r io.ByteReader) uint64 {
	var x uint64
	var s uint
	var b byte
	var err error

	for i := 0; ; i++ {
		if b, err = r.ReadByte(); err != nil {
			return 0
		}
		if b < 0x80 {
			if i > 9 || i == 9 && b > 1 {
				return 0 // overflow
			}
			return x | uint64(b)<<s
		}
		x |= uint64(b&0x7f) << s
		s += 7
	}
}
