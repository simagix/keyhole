// Copyright 2019 Kuei-chun Chen. All rights reserved.

package analytics

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// DocumentDoc contains db.serverStatus().document
type DocumentDoc struct {
	Deleted  int `json:"deleted" bson:"deleted"`
	Inserted int `json:"inserted" bson:"inserted"`
	Returned int `json:"returned" bson:"returned"`
	Updated  int `json:"updated" bson:"updated"`
}

// ExtraInfoDoc contains db.serverStatus().extra_info
type ExtraInfoDoc struct {
	PageFaults uint64 `json:"page_faults" bson:"page_faults"`
}

// GlobalLockSubDoc contains db.serverStatus().globalLockDoc.[activeClients|currentQueue]
type GlobalLockSubDoc struct {
	Readers uint64 `json:"readers" bson:"readers"`
	Total   uint64 `json:"total" bson:"total"`
	Writers uint64 `json:"writers" bson:"writers"`
}

// GlobalLockDoc contains db.serverStatus().globalLockDoc
type GlobalLockDoc struct {
	ActiveClients GlobalLockSubDoc `json:"activeClients" bson:"activeClients"`
	CurrentQueue  GlobalLockSubDoc `json:"currentQueue" bson:"currentQueue"`
	TotalTime     int              `json:"totalTime" bson:"totalTime"`
}

// MemDoc containers db.serverStatus().mem
type MemDoc struct {
	Resident uint64 `json:"resident" bson:"resident"`
	Virtual  uint64 `json:"virtual" bson:"virtual"`
}

// MetricsDoc contains db.serverStatus().metrics
type MetricsDoc struct {
	Document      DocumentDoc      `json:"document" bson:"document"`
	QueryExecutor QueryExecutorDoc `json:"queryExecutor" bson:"queryExecutor"`
	Operation     OperationDoc     `json:"operation" bson:"operation"`
}

// NetworkDoc contains db.serverStatus().network
type NetworkDoc struct {
	BytesIn          uint64 `json:"bytesIn" bson:"bytesIn"`
	BytesOut         uint64 `json:"bytesOut" bson:"bytesOut"`
	NumRequests      uint64 `json:"numRequests" bson:"numRequests"`
	PhysicalBytesIn  uint64 `json:"physicalBytesIn" bson:"physicalBytesIn"`
	PhysicalBytesOut uint64 `json:"physicalBytesOut" bson:"physicalBytesOut"`
}

// OperationDoc contains db.serverStatus().operation
type OperationDoc struct {
	ScanAndOrder   uint64 `json:"scanAndOrder" bson:"scanAndOrder"`
	WriteConflicts uint64 `json:"writeConflicts" bson:"writeConflicts"`
}

// OpCountersDoc contains db.serverStatus().OpCounters
type OpCountersDoc struct {
	Command uint64 `json:"command" bson:"command"`
	Delete  uint64 `json:"delete" bson:"delete"`
	Getmore uint64 `json:"getmore" bson:"getmore"`
	Insert  uint64 `json:"insert" bson:"insert"`
	Query   uint64 `json:"query" bson:"query"`
	Update  uint64 `json:"update" bson:"update"`
}

// OpLatenciesDoc contains db.serverStatus().opLatencies
type OpLatenciesDoc struct {
	Commands OpLatenciesOpDoc `json:"commands" bson:"commands"`
	Reads    OpLatenciesOpDoc `json:"reads" bson:"reads"`
	Writes   OpLatenciesOpDoc `json:"writes" bson:"writes"`
}

// OpLatenciesOpDoc contains doc of db.serverStatus().opLatencies
type OpLatenciesOpDoc struct {
	Latency uint64 `json:"latency" bson:"latency"`
	Ops     uint64 `json:"ops" bson:"ops"`
}

// QueryExecutorDoc contains db.serverStatus().queryExecutor
type QueryExecutorDoc struct {
	Scanned        uint64 `json:"scanned" bson:"scanned"`
	ScannedObjects uint64 `json:"scannedObjects" bson:"scannedObjects"`
}

// WiredTigerBlockManagerDoc contains db.serverStatus().wiredTiger.cache
type WiredTigerBlockManagerDoc struct {
	BytesRead              uint64 `json:"bytes read"`
	BytesWritten           uint64 `json:"bytes written"`
	BytesWrittenCheckPoint uint64 `json:"bytes written for checkpoint"`
}

// WiredTigerCacheDoc contains db.serverStatus().wiredTiger.cache
type WiredTigerCacheDoc struct {
	BytesReadIntoCache     uint64 `json:"bytes read into cache" bson:"bytes read into cache"`
	BytesWrittenFromCache  uint64 `json:"bytes written from cache" bson:"bytes written from cache"`
	CurrentlyInCache       uint64 `json:"bytes currently in the cache" bson:"bytes currently in the cache"`
	MaxBytesConfigured     uint64 `json:"maximum bytes configured" bson:"maximum bytes configured"`
	ModifiedPagesEvicted   uint64 `json:"modified pages evicted" bson:"modified pages evicted"`
	UnmodifiedPagesEvicted uint64 `json:"unmodified pages evicted" bson:"unmodified pages evicted"`
	TrackedDirtyBytes      uint64 `json:"tracked dirty bytes in the cache" bson:"tracked dirty bytes in the cache"`
}

// WiredTigerDataHandleDoc contains db.serverStatus().wiredTiger.data-handle
type WiredTigerDataHandleDoc struct {
	Active uint64 `json:"connection data handles currently active" bson:"connection data handles currently active"`
	Size   uint64 `json:"connection data handle size" bson:"connection data handle size"`
}

// ConcurrentTransactionsCountDoc contains db.serverStatus().wiredTiger.concurrentTransactions.[read|write]
type ConcurrentTransactionsCountDoc struct {
	Available    uint64 `json:"available" bson:"available"`
	Out          uint64 `json:"out" bson:"out"`
	TotalTickets uint64 `json:"totalTickets" bson:"totalTickets"`
}

// ConcurrentTransactionsDoc contains db.serverStatus().wiredTiger.concurrentTransactions
type ConcurrentTransactionsDoc struct {
	Read  ConcurrentTransactionsCountDoc `json:"read" bson:"read"`
	Write ConcurrentTransactionsCountDoc `json:"write" bson:"write"`
}

// WiredTigerDoc contains db.serverStatus().wiredTiger
type WiredTigerDoc struct {
	Perf                   interface{}               `json:"perf" bson:"perf"`
	BlockManager           WiredTigerBlockManagerDoc `json:"block-manager" bson:"block-manager"`
	Cache                  WiredTigerCacheDoc        `json:"cache" bson:"cache"`
	ConcurrentTransactions ConcurrentTransactionsDoc `json:"concurrentTransactions" bson:"concurrentTransactions"`
	DataHandle             WiredTigerDataHandleDoc   `json:"data-handle" bson:"data-handle"`
}

// ConnectionsDoc contains db.serverStatus().connections
type ConnectionsDoc struct {
	Current      uint64 `json:"current" bson:"current"`
	Available    uint64 `json:"available" bson:"available"`
	TotalCreated uint64 `json:"totalCreated" bson:"totalCreated"`
	Active       uint64 `json:"active" bson:"active"`
}

// ServerStatusDoc contains docs from db.serverStatus()
type ServerStatusDoc struct {
	Connections ConnectionsDoc `json:"connections" bson:"connections"`
	ExtraInfo   ExtraInfoDoc   `json:"extra_info" bson:"extra_info"`
	GlobalLock  GlobalLockDoc  `json:"globalLock" bson:"globalLock"`
	Host        string         `json:"host" bson:"host"`
	LocalTime   time.Time      `json:"localTime" bson:"localTime"`
	Mem         MemDoc         `json:"mem" bson:"mem"`
	Metrics     MetricsDoc     `json:"metrics" bson:"metrics"`
	Network     NetworkDoc     `json:"network" bson:"network"`
	OpCounters  OpCountersDoc  `json:"opcounters" bson:"opcounters"`
	OpLatencies OpLatenciesDoc `json:"opLatencies" bson:"opLatencies"`
	Process     string         `json:"process" bson:"process"`
	Repl        bson.M         `json:"repl" bson:"repl"`
	Sharding    bson.M         `json:"sharding" bson:"sharding"`
	Uptime      uint64         `json:"uptime" bson:"uptime"`
	Version     string         `json:"version" bson:"version"`
	WiredTiger  WiredTigerDoc  `json:"wiredTiger" bson:"wiredTiger"`
}
