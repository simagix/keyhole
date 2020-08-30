// Copyright 2019 Kuei-chun Chen. All rights reserved.

package analytics

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/simagix/gox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Metrics stores metrics from FTDC data
type Metrics struct {
	sync.RWMutex
	endpoints []string
	ftdcStats FTDCStats
	latest    int // latest n files
	verbose   bool
}

// FTDCStats FTDC stats
type FTDCStats struct {
	DiskStats         map[string]DiskStats
	MaxWTCache        float64
	ReplicationLags   map[string]TimeSeriesDoc
	ReplSetLegends    []string
	ReplSetStatusList []ReplSetStatusDoc
	ServerInfo        ServerInfoDoc
	ServerStatusList  []ServerStatusDoc
	SystemMetricsList []SystemMetricsDoc
	TimeSeriesData    map[string]TimeSeriesDoc
}

// DiskStats -
type DiskStats struct {
	IOPS         TimeSeriesDoc
	IOInProgress TimeSeriesDoc
	IOQueuedMS   TimeSeriesDoc
	ReadTimeMS   TimeSeriesDoc
	WriteTimeMS  TimeSeriesDoc
	Utilization  TimeSeriesDoc
}

type directoryReq struct {
	Dir  string `json:"dir"`
	Span int    `json:"span"`
}

// NewMetrics returns &Metrics
func NewMetrics() *Metrics {
	gob.Register(primitive.DateTime(1))
	gob.Register(primitive.A{})
	gob.Register(primitive.D{})
	gob.Register(primitive.M{})
	m := Metrics{}
	http.HandleFunc("/grafana", gox.Cors(m.Handler))
	http.HandleFunc("/grafana/", gox.Cors(m.Handler))
	http.HandleFunc("/scores", gox.Cors(m.Handler))
	http.HandleFunc("/scores/", gox.Cors(m.Handler))
	return &m
}

const analyticsEndpoint = `/d/simagix-grafana/mongodb-mongo-ftdc?orgId=1&from=%v&to=%v`
const disksEndpoint = `/d/simagix-grafana-disks/mongodb-disks-stats?orgId=1&from=%v&to=%v`

// SetVerbose sets verbose mode
func (m *Metrics) SetVerbose(verbose bool) { m.verbose = verbose }

// SetLatest sets latest
func (m *Metrics) SetLatest(latest int) { m.latest = latest }

// ProcessFiles reads metrics files/data
func (m *Metrics) ProcessFiles(filenames []string) error {
	hostname, _ := os.Hostname()
	port := 3000
	filenames = GetMetricsFilenames(filenames)
	if len(filenames) == 0 {
		t := time.Now().Unix() * 1000
		minute := int64(60) * 1000
		endpoint := fmt.Sprintf(analyticsEndpoint, t, t+(10*minute))
		log.Println(fmt.Sprintf("http://localhost:%d%v", port, endpoint))
		return errors.New("no available data files found")
	}
	if m.latest > 0 && m.latest < len(filenames) {
		filenames = filenames[len(filenames)-m.latest:]
	}
	if hostname == "ftdc" { // from docker-compose
		port = 3030
		if len(filenames) > 3 { // avoid OOM killer
			fmt.Println("* limits to latest 3 files in a Docker container")
			filenames = filenames[len(filenames)-3:]
		}
	}
	diag := NewDiagnosticData()
	if err := diag.DecodeDiagnosticData(filenames); err != nil { // get summary
		return err
	}
	m.endpoints = diag.endpoints
	m.AddFTDCDetailStats(diag)
	for _, endpoint := range diag.endpoints {
		log.Println(fmt.Sprintf("http://localhost:%d%v", port, endpoint))
	}
	return nil
}

func (m *Metrics) readProcessedFTDC(infile string) error {
	log.Println("Reading from processed FTDC data", infile)
	var err error
	var data []byte
	var file *os.File
	var reader *bufio.Reader

	if file, err = os.Open(infile); err != nil {
		return err
	}
	if reader, err = gox.NewReader(file); err != nil {
		return err
	}
	if data, err = ioutil.ReadAll(reader); err != nil {
		return err
	}
	buffer := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buffer)
	if err = dec.Decode(&m.ftdcStats); err != nil {
		return err
	}
	points := m.ftdcStats.TimeSeriesData["wt_cache_used"].DataPoints
	tm1 := time.Unix(0, int64(points[0][1])*int64(time.Millisecond)).Unix() * 1000
	tm2 := time.Unix(0, int64(points[len(points)-1][1])*int64(time.Millisecond)).Unix() * 1000
	log.Println(tm1, tm2)
	endpoint := fmt.Sprintf(analyticsEndpoint, tm1, tm2)
	log.Printf("http://localhost:3000%v\n", endpoint)
	log.Printf("http://localhost:3030%v\n", endpoint)
	return err
}

// Handler handle HTTP requests
func (m *Metrics) Handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/grafana/query" {
		m.query(w, r)
	} else if r.URL.Path == "/grafana/search" {
		m.search(w, r)
	} else if r.URL.Path == "/grafana/dir" {
		m.readDirectory(w, r)
	} else if strings.HasPrefix(r.URL.Path, "/scores/") {
		fmt.Fprintf(w, GetFormulaHTML(r.URL.Path[9:]))
	} else {
		json.NewEncoder(w).Encode(bson.M{"ok": 1, "message": "hello mongo-ftdc!"})
	}
}

func (m *Metrics) readDirectory(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodOptions:
	case http.MethodPost:
		decoder := json.NewDecoder(r.Body)
		var dr directoryReq
		if err := decoder.Decode(&dr); err != nil {
			json.NewEncoder(w).Encode(bson.M{"ok": 0, "err": err.Error()})
			return
		}
		filenames := getFilenames([]string{dr.Dir})
		if err := m.ProcessFiles(filenames); err != nil {
			json.NewEncoder(w).Encode(bson.M{"ok": 0, "err": err.Error()})
		} else {
			json.NewEncoder(w).Encode(bson.M{"ok": 1, "endpoints": strings.Join(m.endpoints, ",")})
		}
	default:
		http.Error(w, "bad method; supported OPTIONS, POST", http.StatusBadRequest)
		return
	}
}

func (m *Metrics) search(w http.ResponseWriter, r *http.Request) {
	var list []string
	for _, doc := range m.ftdcStats.TimeSeriesData {
		list = append(list, doc.Target)
	}

	list = append(list, "host_info")
	json.NewEncoder(w).Encode(list)
}

func (m *Metrics) query(w http.ResponseWriter, r *http.Request) {
	var tsData []interface{}
	decoder := json.NewDecoder(r.Body)
	var qr QueryRequest
	if err := decoder.Decode(&qr); err != nil {
		json.NewEncoder(w).Encode(tsData)
		return
	}
	ftdc := m.ftdcStats
	for _, target := range qr.Targets {
		if target.Type == "timeserie" {
			if target.Target == "replication_lags" && len(ftdc.ReplicationLags) > 0 { // replaced with actual hostname
				for k, v := range ftdc.ReplicationLags {
					data := v
					data.Target = k
					tsData = append(tsData, FilterTimeSeriesData(data, qr.Range.From, qr.Range.To))
				}
			} else if target.Target == "disks_utils" && len(ftdc.DiskStats) > 0 {
				for k, v := range ftdc.DiskStats {
					data := v.Utilization
					data.Target = k
					tsData = append(tsData, FilterTimeSeriesData(data, qr.Range.From, qr.Range.To))
				}
			} else if target.Target == "disks_iops" && len(ftdc.DiskStats) > 0 {
				for k, v := range ftdc.DiskStats {
					data := v.IOPS
					data.Target = k
					tsData = append(tsData, FilterTimeSeriesData(data, qr.Range.From, qr.Range.To))
				}
			} else if target.Target == "disks_queue_length" && len(ftdc.DiskStats) > 0 {
				for k, v := range ftdc.DiskStats {
					data := v.IOInProgress
					data.Target = k
					tsData = append(tsData, FilterTimeSeriesData(data, qr.Range.From, qr.Range.To))
				}
			} else if target.Target == "read_time_ms" && len(ftdc.DiskStats) > 0 {
				for k, v := range ftdc.DiskStats {
					data := v.ReadTimeMS
					data.Target = k
					tsData = append(tsData, FilterTimeSeriesData(data, qr.Range.From, qr.Range.To))
				}
			} else if target.Target == "write_time_ms" && len(ftdc.DiskStats) > 0 {
				for k, v := range ftdc.DiskStats {
					data := v.WriteTimeMS
					data.Target = k
					tsData = append(tsData, FilterTimeSeriesData(data, qr.Range.From, qr.Range.To))
				}
			} else if target.Target == "io_queued_ms" && len(ftdc.DiskStats) > 0 {
				for k, v := range ftdc.DiskStats {
					data := v.IOQueuedMS
					data.Target = k
					tsData = append(tsData, FilterTimeSeriesData(data, qr.Range.From, qr.Range.To))
				}
			} else {
				data := ftdc.TimeSeriesData[target.Target]
				data.Target = GetShortLabel(target.Target)
				tsData = append(tsData, FilterTimeSeriesData(data, qr.Range.From, qr.Range.To))
			}
		} else if target.Type == "table" {
			if target.Target == "host_info" {
				headerList := []bson.M{}
				rowList := [][]string{}
				headerList = append(headerList, bson.M{"text": "Configurations", "type": "String"})
				rowList = append(rowList, []string{fmt.Sprintf(`CPU: %v cores (%v)`,
					m.ftdcStats.ServerInfo.HostInfo.System.NumCores,
					m.ftdcStats.ServerInfo.HostInfo.System.CPUArch)})
				if m.verbose == true {
					rowList = append(rowList, []string{fmt.Sprintf(`Host: %v`, m.ftdcStats.ServerInfo.HostInfo.System.Hostname)})
				}
				rowList = append(rowList, []string{fmt.Sprintf(`Memory: %v`,
					gox.GetStorageSize(1024*1024*m.ftdcStats.ServerInfo.HostInfo.System.MemSizeMB))})
				rowList = append(rowList, []string{m.ftdcStats.ServerInfo.HostInfo.OS.Type + " (" + m.ftdcStats.ServerInfo.HostInfo.OS.Version + ")"})
				rowList = append(rowList, []string{m.ftdcStats.ServerInfo.HostInfo.OS.Name})
				rowList = append(rowList, []string{fmt.Sprintf(`MongoDB v%v`, m.ftdcStats.ServerInfo.BuildInfo.Version)})
				doc := bson.M{"columns": headerList, "type": "table", "rows": rowList}
				tsData = append(tsData, doc)
			} else if target.Target == "assessment" {
				as := NewAssessment(ftdc)
				as.SetVerbose(m.verbose)
				tsData = append(tsData, as.GetAssessment(qr.Range.From, qr.Range.To))
			}
		}
	}
	json.NewEncoder(w).Encode(tsData)
}

// AddFTDCDetailStats assign FTDC values
func (m *Metrics) AddFTDCDetailStats(diag *DiagnosticData) {
	m.RLock()
	defer m.RUnlock()
	ftdc := &m.ftdcStats

	sort.Slice(diag.ReplSetStatusList, func(i int, j int) bool {
		return diag.ReplSetStatusList[i].Date.Before(diag.ReplSetStatusList[j].Date)
	})
	if len(ftdc.ReplSetStatusList) == 0 {
		ftdc.ReplSetStatusList = diag.ReplSetStatusList
	} else {
		lastOne := ftdc.ReplSetStatusList[len(ftdc.ReplSetStatusList)-1]
		for i, v := range diag.ReplSetStatusList {
			if v.Date.After(lastOne.Date) {
				ftdc.ReplSetStatusList = append(ftdc.ReplSetStatusList, diag.ReplSetStatusList[i:]...)
				break
			}
		}
	}

	sort.Slice(diag.ServerStatusList, func(i int, j int) bool {
		return diag.ServerStatusList[i].LocalTime.Before(diag.ServerStatusList[j].LocalTime)
	})
	if len(ftdc.ServerStatusList) == 0 {
		ftdc.ServerStatusList = diag.ServerStatusList
	} else {
		lastOne := ftdc.ServerStatusList[len(ftdc.ServerStatusList)-1]
		for i, v := range diag.ServerStatusList {
			if v.LocalTime.After(lastOne.LocalTime) {
				ftdc.ServerStatusList = append(ftdc.ServerStatusList, diag.ServerStatusList[i:]...)
				break
			}
		}
	}

	sort.Slice(diag.SystemMetricsList, func(i int, j int) bool {
		return diag.SystemMetricsList[i].Start.Before(diag.SystemMetricsList[j].Start)
	})
	if len(ftdc.SystemMetricsList) == 0 {
		ftdc.SystemMetricsList = diag.SystemMetricsList
	} else {
		lastOne := ftdc.SystemMetricsList[len(ftdc.SystemMetricsList)-1]
		for i, v := range diag.SystemMetricsList {
			if v.Start.After(lastOne.Start) {
				ftdc.SystemMetricsList = append(ftdc.SystemMetricsList, diag.SystemMetricsList[i:]...)
				break
			}
		}
	}

	b, _ := json.Marshal(diag.ServerInfo)
	btm := time.Now()
	var wiredTigerTSD map[string]TimeSeriesDoc
	var replicationTSD map[string]TimeSeriesDoc
	var systemMetricsTSD map[string]TimeSeriesDoc

	var wg = gox.NewWaitGroup(4) // use 4 threads to read
	wg.Add(1)
	go func() {
		defer wg.Done()
		replicationTSD, ftdc.ReplicationLags = getReplSetGetStatusTimeSeriesDoc(ftdc.ReplSetStatusList, &ftdc.ReplSetLegends) // replSetGetStatus
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		systemMetricsTSD, ftdc.DiskStats = getSystemMetricsTimeSeriesDoc(ftdc.SystemMetricsList) // SystemMetrics
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		ftdc.TimeSeriesData = getServerStatusTimeSeriesDoc(ftdc.ServerStatusList) // ServerStatus
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		wiredTigerTSD = getWiredTigerTimeSeriesDoc(ftdc.ServerStatusList) // ServerStatus
	}()
	wg.Wait()

	// merge
	for k, v := range wiredTigerTSD {
		ftdc.TimeSeriesData[k] = v
	}
	for k, v := range replicationTSD {
		ftdc.TimeSeriesData[k] = v
	}
	for k, v := range systemMetricsTSD {
		ftdc.TimeSeriesData[k] = v
	}
	json.Unmarshal(b, &ftdc.ServerInfo)
	if len(m.ftdcStats.TimeSeriesData["wt_cache_max"].DataPoints) > 0 && len(m.ftdcStats.TimeSeriesData["wt_cache_max"].DataPoints[0]) > 0 {
		m.ftdcStats.MaxWTCache = m.ftdcStats.TimeSeriesData["wt_cache_max"].DataPoints[0][0]
	}
	etm := time.Now()
	if m.verbose == true {
		log.Println("data points added for", m.ftdcStats.ServerInfo.HostInfo.System.Hostname, ", time spent:", etm.Sub(btm).String())
	}
}

// FilterTimeSeriesData returns partial data points if there are too many
func FilterTimeSeriesData(tsData TimeSeriesDoc, from time.Time, to time.Time) TimeSeriesDoc {
	seconds := 1800.0 // .5 hour, no gain to have a higher number.  Grafana aggregates
	if len(tsData.DataPoints) == 0 {
		return tsData
	}
	var data = TimeSeriesDoc{Target: tsData.Target, DataPoints: [][]float64{}}
	fidx := findClosestDataPointIndex(tsData.DataPoints, float64(from.UnixNano()/1000000))
	eidx := findClosestDataPointIndex(tsData.DataPoints, float64(to.UnixNano()/1000000))
	points := tsData.DataPoints[fidx:eidx]
	if len(points) == 0 || math.IsNaN(points[0][0]) {
		return data
	}

	if len(points) > int(seconds) {
		length := float64(len(points) + 1)
		samples := [][]float64{}
		for i := 0.0; i < seconds; i++ { // for fast sampling
			p := int(length * i / seconds)
			samples = append(samples, points[p])
		}
		data.DataPoints = samples
	} else {
		data.DataPoints = points
	}
	return data
}

// perform binary search
func findClosestDataPointIndex(arr [][]float64, target float64) int {
	n := len(arr)
	if target <= arr[0][1] {
		return 0
	}
	if target >= arr[n-1][1] {
		return n - 1
	}
	i := 0
	j := n
	mid := 0
	for i < j {
		mid = (i + j) / 2
		if arr[mid][1] == target {
			return mid
		}
		if target < arr[mid][1] {
			if mid > 0 && target > arr[mid-1][1] {
				if target-arr[mid-1][1] >= arr[mid][1] {
					return mid
				}
				return mid - 1
			}
			j = mid
		} else {
			if mid < n-1 && target < arr[mid+1][1] {
				if target-arr[mid-1][1] >= arr[mid][1] {
					return mid
				}
				return mid - 1
			}
			i = mid + 1
		}
	}
	return mid
}
