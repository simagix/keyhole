// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/globalsign/mgo/bson"
)

func (g *Grafana) handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path[1:] == "grafana/" {
		fmt.Fprintf(w, "ok\n")
	} else if r.URL.Path[1:] == "grafana/query" {
		g.query(w, r)
	} else if r.URL.Path[1:] == "grafana/search" {
		g.search(w, r)
	} else if r.URL.Path[1:] == "grafana/dir" {
		g.readDirectory(w, r)
	}
}

type directoryReq struct {
	Dir  string `json:"dir"`
	Span int    `json:"span"`
}

func (g *Grafana) readDirectory(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodOptions:
	case http.MethodPost:
		var err error
		decoder := json.NewDecoder(r.Body)
		var dr directoryReq
		if err = decoder.Decode(&dr); err != nil {
			json.NewEncoder(w).Encode(bson.M{"ok": 0, "err": err.Error()})
			return
		}
		if dr.Span < 1 {
			dr.Span = 1
		}
		d := NewDiagnosticData(dr.Span)
		var filenames = []string{dr.Dir}
		if _, err = d.PrintDiagnosticData(filenames, true); err != nil {
			json.NewEncoder(w).Encode(bson.M{"ok": 0, "err": err.Error()})
			return
		}
		g.ReinitGrafana(d)
		json.NewEncoder(w).Encode(bson.M{"ok": 1, "dir": dr.Dir})
	default:
		http.Error(w, "bad method; supported OPTIONS, POST", http.StatusBadRequest)
		return
	}
}

func (g *Grafana) search(w http.ResponseWriter, r *http.Request) {
	var list []string

	for _, doc := range g.timeSeriesData {
		list = append(list, doc.Target)
	}

	list = append(list, "host_info")
	json.NewEncoder(w).Encode(list)
}

func (g *Grafana) query(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var qr QueryRequest
	if err := decoder.Decode(&qr); err != nil {
		return
	}

	var tsData []interface{}
	for _, target := range qr.Targets {
		if target.Type == "timeserie" {
			if target.Target == "replication_lags" { // replaced with actual hostname
				for k, v := range g.replicationLags {
					data := v
					data.Target = k
					tsData = append(tsData, filterTimeSeriesData(data, qr.Range.From, qr.Range.To))
				}
			} else if target.Target == "disks_utils" {
				for k, v := range g.diskStats {
					data := v.utilization
					data.Target = k
					tsData = append(tsData, filterTimeSeriesData(data, qr.Range.From, qr.Range.To))
				}
			} else if target.Target == "disks_iops" {
				for k, v := range g.diskStats {
					data := v.iops
					data.Target = k
					tsData = append(tsData, filterTimeSeriesData(data, qr.Range.From, qr.Range.To))
				}
			} else {
				tsData = append(tsData, filterTimeSeriesData(g.timeSeriesData[target.Target], qr.Range.From, qr.Range.To))
			}
		} else if target.Type == "table" {
			if target.Target == "host_info" {
				headerList := []bson.M{}
				headerList = append(headerList, bson.M{"text": "Info", "type": "string"})
				headerList = append(headerList, bson.M{"text": "Value", "type": "string"})
				var si ServerInfoDoc
				b, _ := json.Marshal(g.serverInfo)
				if err := json.Unmarshal(b, &si); err != nil {
					rowList := [][]string{[]string{"Error", err.Error()}}
					doc1 := bson.M{"columns": headerList, "type": "table", "rows": rowList}
					tsData = append(tsData, doc1)
					continue
				}
				rowList := [][]string{}

				rowList = append(rowList, []string{"CPU", strconv.Itoa(si.HostInfo.System.NumCores) + " cores (" + si.HostInfo.System.CPUArch + ")"})
				rowList = append(rowList, []string{"Hostname", si.HostInfo.System.Hostname})
				rowList = append(rowList, []string{"Memory (MB)", strconv.Itoa(si.HostInfo.System.MemSizeMB)})
				rowList = append(rowList, []string{"MongoDB Version", si.BuildInfo.Version})
				rowList = append(rowList, []string{"OS", si.HostInfo.OS.Name})
				rowList = append(rowList, []string{"OS Type", si.HostInfo.OS.Type + " (" + si.HostInfo.OS.Version + ")"})
				doc1 := bson.M{"columns": headerList, "type": "table", "rows": rowList}
				tsData = append(tsData, doc1)
			}
		}
	}
	json.NewEncoder(w).Encode(tsData)
}

func filterTimeSeriesData(tsData TimeSeriesDoc, from time.Time, to time.Time) TimeSeriesDoc {
	var data = TimeSeriesDoc{DataPoints: [][]float64{}}
	data.Target = tsData.Target
	for _, v := range tsData.DataPoints {
		tm := time.Unix(0, int64(v[1])*int64(time.Millisecond))
		if tm.After(to) || tm.Before(from) {
			continue
		}
		data.DataPoints = append(data.DataPoints, v)
	}

	max := 600 // at most 10 minutes
	if len(data.DataPoints) > max {
		frac := len(data.DataPoints) / max
		var datax = TimeSeriesDoc{DataPoints: [][]float64{}}
		datax.Target = tsData.Target
		for i, v := range data.DataPoints {
			if i%frac != 0 {
				continue
			}
			datax.DataPoints = append(datax.DataPoints, v)
		}
		return datax
	}
	return data
}
