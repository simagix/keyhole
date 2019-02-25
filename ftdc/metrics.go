// Copyright 2018 Kuei-chun Chen. All rights reserved.

package ftdc

import (
	"bytes"
	"compress/zlib"
	"io"
	"io/ioutil"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ReadMetricsSummary read summary (first doc) of metrics
func (m *Metrics) ReadMetricsSummary(buffer []byte) error {
	return m.readMetrics(buffer, true)
}

// ReadAllMetrics reads all metrics
func (m *Metrics) ReadAllMetrics(buffer []byte) error {
	return m.readMetrics(buffer, false)
}

// readMetrics reads FTDC metrics from bytes buffer
func (m *Metrics) readMetrics(buffer []byte, summaryOnly bool) error {
	var err error
	var pos uint32
	var metricsData = []MetricsData{}
	var md MetricsData
	var r io.ReadCloser

	for {
		if pos >= uint32(len(buffer)) {
			break
		}
		bs := buffer[pos:(pos + 4)]
		var length = GetUint32(bytes.NewReader(bs))
		bs = buffer[pos:(pos + length)]
		pos += length
		var out = bson.M{}
		if err = bson.Unmarshal(bs, &out); err != nil {
			return err
		} else if out["type"] == int32(0) {
			m.Doc = out["doc"]
		} else if out["type"] == int32(1) {
			bytesBuf := bytes.NewReader((out["data"].(primitive.Binary)).Data[4:])
			// zlib decompress
			if r, err = zlib.NewReader(bytesBuf); err != nil {
				return err
			}
			var block []byte
			if block, err = ioutil.ReadAll(r); err != nil {
				return err
			}

			if summaryOnly == true {
				r := bytes.NewReader(block)
				md = MetricsData{DataPointsMap: map[string][]int64{}, Buffer: block, DocSize: GetUint32(r)}
				metricsData = append(metricsData, md)
			} else {
				if md, err = m.decode(block); err != nil {
					return err
				}
				metricsData = append(metricsData, md)
			}
		} else {
			// log.Println("==>", out["type"])
		}
	}
	m.Data = metricsData
	return err
}
