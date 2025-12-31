// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"

	"github.com/simagix/gox"
	"go.mongodb.org/mongo-driver/mongo"
)

// Sharded cluster
const Sharded = "sharded"

// Replica set
const Replica = "replica"

// Standalone server
const Standalone = "standalone"

// ClusterStats keeps slow ops struct
type ClusterStats struct {
	BuildInfo        BuildInfo        `bson:"buildInfo"`
	CmdLineOpts      CmdLineOpts      `bson:"getCmdLineOpts"`
	Cluster          string           `bson:"cluster"`
	Databases        *[]Database      `bson:"databases"`
	Host             string           `bson:"host"`
	HostInfo         HostInfo         `bson:"hostInfo"`
	Logger           *gox.Logger      `bson:"keyhole"`
	OplogStats       OplogStats       `bson:"oplog"`
	Process          string           `bson:"process"`
	ReplSetGetStatus ReplSetGetStatus `bson:"replSetGetStatus"`
	ServerStatus     ServerStatus     `bson:"serverStatus"`
	Shards           []Shard          `bson:"shards"`
	Version          string           `bson:"version"`

	dbNames   []string
	fastMode  bool
	redact    bool
	signature string
	verbose   bool
}

// NewClusterStats returns *ClusterStats
func NewClusterStats(signature string) *ClusterStats {
	s := ClusterStats{signature: signature}
	return &s
}

// SetDBNames sets redact
func (p *ClusterStats) SetDBNames(dbNames []string) {
	p.dbNames = dbNames
}

// SetFastMode sets fastMode mode
func (p *ClusterStats) SetFastMode(fastMode bool) {
	p.fastMode = fastMode
}

// SetRedaction sets redact
func (p *ClusterStats) SetRedaction(redact bool) {
	p.redact = redact
}

// SetVerbose sets verbose mode
func (p *ClusterStats) SetVerbose(verbose bool) {
	p.verbose = verbose
}

// GetClusterStats collects cluster stats
func (p *ClusterStats) GetClusterStats(client *mongo.Client, connString connstring.ConnString) error {
	var err error
	p.Logger = gox.GetLogger(p.signature)
	p.Logger.Debug("GetClusterStats() begins")
	if err = p.GetClusterStatsSummary(client); err != nil {
		return err
	}
	if p.CmdLineOpts, err = GetCmdLineOpts(client); err != nil {
		p.Logger.Info(fmt.Sprintf(`GetCmdLineOpts(): %v`, err))
	}
	if p.Cluster == Sharded { //collects from the primary of each shard
		if p.Shards, err = GetShards(client); err != nil {
			p.Logger.Infof(`GetShards(): %v`, err)
		}
		p.Logger.Infof("%v shards detected, collecting from all servers", len(p.Shards))
		if p.Shards, err = p.GetServersStatsSummary(p.Shards, connString); err != nil {
			p.Logger.Error(err)
		}
		p.Logger.Info("end collecting from all servers")
	} else if p.Cluster == Replica && p.Process == "mongod" { //collects replica info
		message := "replica detected, collecting from all servers"
		p.Logger.Info(message)
		if p.ReplSetGetStatus, err = GetReplSetGetStatus(client); err != nil {
			p.Logger.Info(fmt.Sprintf(`GetReplSetGetStatus(): %v`, err))
		}

		setName := p.ServerStatus.Repl.SetName
		hosts := []string{}
		for _, member := range p.ReplSetGetStatus.Members {
			hosts = append(hosts, member.Name)
		}
		s := fmt.Sprintf(`%v/%v`, setName, strings.Join(hosts, ","))
		oneShard := []Shard{{ID: setName, State: 1, Host: s}}
		if p.Shards, err = p.GetServersStatsSummary(oneShard, connString); err != nil {
			p.Logger.Error(err)
		}
		p.Logger.Info("end collecting from all servers")
	}
	db := NewDatabaseStats(p.Logger.AppName)
	db.SetNumberShards(len(p.Shards))
	db.SetRedaction(p.redact)
	db.SetVerbose(p.verbose)
	db.SetFastMode(p.fastMode)
	var databases []Database
	if databases, err = db.GetAllDatabasesStats(client, p.dbNames); err != nil {
		p.Logger.Info(fmt.Sprintf(`GetAllDatabasesStats(): %v`, err))
	}
	p.Databases = &databases
	return nil
}

// getShortErrorMessage extracts a concise error message from verbose MongoDB errors
func getShortErrorMessage(err error) string {
	msg := err.Error()
	// Extract just the error code and reason, e.g., "(AtlasError) (Unauthorized)"
	if idx := strings.Index(msg, " not authorized"); idx > 0 {
		return msg[:idx]
	}
	if idx := strings.Index(msg, " command "); idx > 0 {
		return msg[:idx]
	}
	// Truncate very long messages
	if len(msg) > 80 {
		return msg[:80] + "..."
	}
	return msg
}

// GetClusterStatsSummary collects cluster stats
// Commands that fail (e.g., on Atlas M0 shared clusters) are logged and skipped
func (p *ClusterStats) GetClusterStatsSummary(client *mongo.Client) error {
	var err error
	var isSharedCluster bool
	p.Logger = gox.GetLogger(p.signature)
	p.Logger.Debug("GetBuildInfo()")
	if p.BuildInfo, err = GetBuildInfo(client); err != nil {
		return err // buildInfo is required, fail if not available
	}
	p.Version = p.BuildInfo.Version
	p.Logger.Debug("GetHostInfo()")
	if p.HostInfo, err = GetHostInfo(client); err != nil {
		// hostInfo fails on shared clusters (M0/M2/M5), mark as shared and continue
		isSharedCluster = true
		p.Logger.Warnf("GetHostInfo() skipped: %v", getShortErrorMessage(err))
	}
	p.Logger.Debug("GetServerStatus()")
	if p.ServerStatus, err = GetServerStatus(client); err != nil {
		// serverStatus may be restricted on shared clusters, continue without it
		p.Logger.Warnf("GetServerStatus() skipped: %v", getShortErrorMessage(err))
	}
	p.Host = p.ServerStatus.Host
	if p.Host == "" && p.HostInfo.System.Hostname != "" {
		p.Host = p.HostInfo.System.Hostname
	}
	p.Process = p.ServerStatus.Process
	p.Cluster = GetClusterType(p.ServerStatus)
	if p.Cluster == Replica && p.Process == "mongod" { //collects replica info
		if isSharedCluster {
			// Skip oplog and replset status on shared clusters - local db is not accessible
			p.Logger.Warn("GetOplogStats() skipped: shared cluster")
			p.Logger.Warn("GetReplSetGetStatus() skipped: shared cluster")
		} else {
			p.Logger.Debug("GetOplogStats()")
			if p.OplogStats, err = GetOplogStats(client); err != nil {
				p.Logger.Warnf("GetOplogStats() skipped: %v", getShortErrorMessage(err))
			}
			p.Logger.Debug("GetReplSetGetStatus()")
			if p.ReplSetGetStatus, err = GetReplSetGetStatus(client); err != nil {
				p.Logger.Warnf("GetReplSetGetStatus() skipped: %v", getShortErrorMessage(err))
			}
		}
	} else if p.Cluster == Sharded {
		p.Logger.Debug("GetShards()")
		if p.Shards, err = GetShards(client); err != nil {
			p.Logger.Warnf("GetShards() skipped: %v", getShortErrorMessage(err))
		}
	}
	return nil
}

// GetServersStatsSummary returns cluster stats from all shards
func (p *ClusterStats) GetServersStatsSummary(shards []Shard, connString connstring.ConnString) ([]Shard, error) {
	var err error
	var uris []string
	var smap = map[string]Shard{}
	p.Logger = gox.GetLogger(p.signature)
	for _, v := range shards {
		v.Servers = []ClusterStats{}
		smap[v.ID] = v
	}
	if uris, err = GetAllServerURIs(shards, connString); err != nil {
		return shards, err
	}
	wg := gox.NewWaitGroup(6)
	var mu sync.Mutex
	var e error
	for i, uri := range uris {
		s := uri
		cs, _ := connstring.Parse(s)
		if cs.Password != "" {
			s = strings.Replace(s, url.QueryEscape(cs.Password), "xxxxxx", 1)
		}
		p.Logger.Infof(`[t-%d] collect from %v`, i, s)
		wg.Add(1)
		go func(uri string, n int, logger *gox.Logger) {
			defer wg.Done()
			var sclient *mongo.Client
			var err error
			if sclient, err = NewMongoClient(uri); err != nil {
				logger.Errorf(`[t-%d] error: %v`, n, err)
				mu.Lock()
				e = err
				mu.Unlock()
				return
			}
			tm := time.Now()
			sclient.Ping(context.Background(), nil)
			logger.Infof(`[t-%d] ping: %v`, n, time.Since(tm))
			defer sclient.Disconnect(context.Background())
			server := NewClusterStats(p.Logger.AppName)
			if err = server.GetClusterStatsSummary(sclient); err != nil {
				logger.Errorf(`[t-%d] error: %v`, n, err)
				mu.Lock()
				e = err
				mu.Unlock()
				return
			}
			mu.Lock()
			node := smap[server.ServerStatus.Repl.SetName]
			node.Servers = append(node.Servers, *server)
			smap[server.ServerStatus.Repl.SetName] = node
			mu.Unlock()
			logger.Infof(`[t-%d] completed`, n)
		}(uri, i, p.Logger)
	}
	wg.Wait()
	if e != nil {
		return shards, e
	}
	shards = []Shard{}
	for _, v := range smap {
		shards = append(shards, v)
	}
	return shards, nil
}

// GetClusterShortSummary returns one line summary
// Returns partial info if some commands fail (e.g., on Atlas M0 shared clusters)
func (p *ClusterStats) GetClusterShortSummary(client *mongo.Client) string {
	var err error
	if err = p.GetClusterStatsSummary(client); err != nil {
		// Only buildInfo failure returns error, other failures are logged and skipped
		return fmt.Sprintf("error: %v", err)
	}
	return p.GetShortSummary()
}

// GetShortSummary returns a short summary
func (p *ClusterStats) GetShortSummary() string {
	edition := "community"
	if len(p.BuildInfo.Modules) > 0 {
		edition = p.BuildInfo.Modules[0]
	}
	numShardStr := ""
	if p.Cluster == Sharded {
		numShardStr = fmt.Sprintf(`(%v)`, len(p.Shards))
	}
	// Handle missing data gracefully (e.g., on Atlas M0 shared clusters)
	hostname := p.HostInfo.System.Hostname
	if hostname == "" {
		hostname = p.Host // fallback to serverStatus.host
	}
	if hostname == "" {
		hostname = "unknown"
	}
	osName := p.HostInfo.OS.Name
	if osName == "" {
		osName = "N/A"
	}
	process := p.ServerStatus.Process
	if process == "" {
		process = "mongod"
	}
	cluster := p.Cluster
	if cluster == "" {
		cluster = "unknown"
	}
	numCores := p.HostInfo.System.NumCores
	memSizeMB := p.HostInfo.System.MemSizeMB
	// Build result with available info
	if numCores > 0 && memSizeMB > 0 {
		return fmt.Sprintf(`MongoDB v%v %v %v (%v) %v %v%v %v cores %v MB mem`,
			p.BuildInfo.Version, edition, hostname, osName,
			process, cluster, numShardStr, numCores, memSizeMB)
	}
	// Minimal output when host info is not available (shared clusters)
	return fmt.Sprintf(`MongoDB v%v %v %v %v %v%v`,
		p.BuildInfo.Version, edition, hostname, process, cluster, numShardStr)
}

// Print prints a cluster short summary
func (p *ClusterStats) Print() {
	fmt.Println(p.GetShortSummary())
}

// OutputBSON writes bson data to a file
// Streams directly to gzip file to minimize memory usage
func (p *ClusterStats) OutputBSON() (string, []byte, error) {
	var err error
	var ofile string
	// Use hostname from hostInfo, fallback to serverStatus.Host (for shared clusters like M0)
	basename := p.HostInfo.System.Hostname
	if basename == "" {
		basename = p.Host
	}
	if basename == "" {
		result := `unable to determine hostname - roles 'clusterMonitor' and 'readAnyDatabase' may be required`
		return ofile, nil, errors.New(result)
	}

	os.Mkdir(outdir, 0755)
	basename = strings.ReplaceAll(basename, ":", "_")
	ofile = fmt.Sprintf(`%v/%v-stats.bson.gz`, outdir, basename)
	i := 1
	for DoesFileExist(ofile) {
		ofile = fmt.Sprintf(`%v/%v.%d-stats.bson.gz`, outdir, basename, i)
		i++
	}

	// Create file and gzip writer for streaming
	file, err := os.Create(ofile)
	if err != nil {
		return ofile, nil, err
	}
	gzWriter := gzip.NewWriter(file)

	// Prepare database summaries (without collections) for ClusterStats header
	databases := p.Databases
	p.Databases = nil
	var summaries []Database
	for _, db := range *databases {
		dbSummary := Database{
			Name:       db.Name,
			SizeOnDisk: db.SizeOnDisk,
			Empty:      db.Empty,
			Shards:     db.Shards,
			Stats:      db.Stats}
		summaries = append(summaries, dbSummary)
	}
	p.Databases = &summaries

	// Stream ClusterStats header directly to gzip
	data, err := bson.Marshal(p)
	if err != nil {
		gzWriter.Close()
		file.Close()
		os.Remove(ofile)
		return ofile, nil, err
	}
	if _, err = gzWriter.Write(data); err != nil {
		gzWriter.Close()
		file.Close()
		os.Remove(ofile)
		return ofile, nil, err
	}

	// Stream each collection directly to gzip
	for _, db := range *databases {
		for _, coll := range db.Collections {
			data, err = bson.Marshal(coll)
			if err != nil {
				gzWriter.Close()
				file.Close()
				os.Remove(ofile)
				return ofile, nil, err
			}
			if _, err = gzWriter.Write(data); err != nil {
				gzWriter.Close()
				file.Close()
				os.Remove(ofile)
				return ofile, nil, err
			}
		}
	}

	// Close gzip writer and file
	if err = gzWriter.Close(); err != nil {
		file.Close()
		os.Remove(ofile)
		return ofile, nil, err
	}
	if err = file.Close(); err != nil {
		return ofile, nil, err
	}

	fmt.Printf("bson data written to %v\n", ofile)

	// Read and decompress file for return value (Maobi expects uncompressed BSON)
	file, err = os.Open(ofile)
	if err != nil {
		return ofile, nil, err
	}
	defer file.Close()
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return ofile, nil, err
	}
	defer gzReader.Close()
	data, err = io.ReadAll(gzReader)
	if err != nil {
		return ofile, nil, err
	}
	return ofile, data, nil
}
