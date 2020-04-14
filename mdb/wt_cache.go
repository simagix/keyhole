// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/simagix/gox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// WiredTigerCache stores wiredTiger cache structure
type WiredTigerCache struct {
	client    *mongo.Client
	databases []bson.M
	numPoints int
}

// NewWiredTigerCache returns *WiredTigerCache
func NewWiredTigerCache(client *mongo.Client) *WiredTigerCache {
	wtc := WiredTigerCache{client: client, numPoints: 10}
	return &wtc
}

// Start starts a web server and a thread to collect caches
func (wtc *WiredTigerCache) Start(port int) {
	var err error
	http.HandleFunc("/", gox.Cors(wtc.Handler))
	http.HandleFunc("/wt", gox.Cors(wtc.Handler))
	http.HandleFunc("/wt/", gox.Cors(wtc.Handler))
	hostname, _ := os.Hostname()
	log.Println(fmt.Sprintf("HTTP server ready, URL: http://%s:%d/", hostname, port))
	go func() {
		log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
	}()
	for {
		if wtc.databases, err = GetAllDatabasesInfo(wtc.client); err != nil {
			log.Fatal(err)
		}
		time.Sleep(5 * time.Second)
	}
}

// Handler supports resetful calls
func (wtc *WiredTigerCache) Handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path[1:] == "wt/data" {
		wtc.GetWiredTigerCacheData(w, r)
	} else if r.URL.Path[1:] == "wt" || r.URL.Path[1:] == "wt/" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))
	} else {
		json.NewEncoder(w).Encode(bson.M{"ok": 1, "message": "hello keyhole!"})
	}
}

// GetWiredTigerCacheData gets WT cache data
func (wtc *WiredTigerCache) GetWiredTigerCacheData(w http.ResponseWriter, r *http.Request) {
	topCaches := []ChartDataPoint{}
	topDataCache := []ChartDataPoint{}
	topIndexesCache := []ChartDataPoint{}
	cacheDataSize := int64(0)
	cacheIndexesSize := int64(0)

	for _, database := range wtc.databases {
		collections := toArray(database["collections"])
		for _, collection := range collections {
			ns := collection["NS"].(string)
			// top storage list
			stats := collection["stats"].(bson.M)
			if stats["wiredTiger"] != nil {
				x := toInt64(stats["wiredTiger"].(bson.M)["cache"].(bson.M)["bytes currently in the cache"])
				cacheDataSize += x
				topCaches = append(topCaches, ChartDataPoint{label: "D:" + ns, value: x})
				if len(topDataCache) < wtc.numPoints {
					topDataCache = append(topDataCache, ChartDataPoint{label: ns, value: x})
				} else if int64(x) > topDataCache[wtc.numPoints-1].value {
					topDataCache[wtc.numPoints-1] = ChartDataPoint{label: ns, value: int64(x)}
				}
				sort.Slice(topDataCache, func(i int, j int) bool {
					return topDataCache[i].value > topDataCache[j].value
				})
			}
			if stats["indexDetails"] != nil {
				indexDetails := stats["indexDetails"].(bson.M)
				x := int64(0)
				for _, v := range indexDetails {
					if v.(bson.M)["cache"] != nil {
						x += toInt64(v.(bson.M)["cache"].(bson.M)["bytes currently in the cache"])
					}
				}
				cacheIndexesSize += x
				topCaches = append(topCaches, ChartDataPoint{label: "I:" + ns, value: x})
				if len(topIndexesCache) < wtc.numPoints {
					topIndexesCache = append(topIndexesCache, ChartDataPoint{label: ns, value: x})
				} else if int64(x) > topIndexesCache[wtc.numPoints-1].value {
					topIndexesCache[wtc.numPoints-1] = ChartDataPoint{label: ns, value: int64(x)}
				}
				sort.Slice(topIndexesCache, func(i int, j int) bool {
					return topIndexesCache[i].value > topIndexesCache[j].value
				})
			}

			if len(topCaches) > 100 {
				sort.Slice(topCaches, func(i int, j int) bool {
					return topCaches[i].value > topCaches[j].value
				})
				topCaches = topCaches[:80]
			}
		}
	}

	// top cache usages
	sort.Slice(topCaches, func(i int, j int) bool {
		return topCaches[i].value > topCaches[j].value
	})
	var unit string
	unit, topCaches = getTopChartPoints(topCaches, wtc.numPoints)
	data := [][]interface{}{[]interface{}{"Name Space", unit}}
	title := fmt.Sprintf("Data and Indexes in WiredTiger Cache (%s)", unit)
	for _, v := range topCaches {
		data = append(data, []interface{}{v.label, v.value})
	}
	doc := bson.M{"top_caches": bson.M{"title": title, "data": data}}

	// cache distr, indexes vs. data
	points := []ChartDataPoint{}
	points = append(points, ChartDataPoint{label: "Data", value: cacheDataSize})
	points = append(points, ChartDataPoint{label: "Indexes", value: cacheIndexesSize})
	unit, points = getTopChartPoints(points, wtc.numPoints)
	title = fmt.Sprintf("Data vs. Indexes in WiredTiger Cache (%s)", unit)
	data = [][]interface{}{[]interface{}{"Name Space", unit}}
	for _, v := range points {
		data = append(data, []interface{}{v.label, v.value})
	}
	doc["cache_distr"] = bson.M{"title": title, "data": data}

	// top data cache
	unit, topDataCache = getTopChartPoints(topDataCache, wtc.numPoints)
	data = [][]interface{}{[]interface{}{"Name Space", unit}}
	title = fmt.Sprintf("Data in WiredTiger Cache (%s)", unit)
	for _, v := range topDataCache {
		data = append(data, []interface{}{v.label, v.value})
	}
	doc["top_data_cache"] = bson.M{"title": title, "data": data}

	// top indexes cache
	unit, topIndexesCache = getTopChartPoints(topIndexesCache, wtc.numPoints)
	data = [][]interface{}{[]interface{}{"Name Space", unit}}
	title = fmt.Sprintf("Indexes in WiredTiger Cache (%s)", unit)
	for _, v := range topIndexesCache {
		data = append(data, []interface{}{v.label, v.value})
	}
	doc["top_indexes_cache"] = bson.M{"title": title, "data": data}
	json.NewEncoder(w).Encode(doc)
}
