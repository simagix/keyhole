// Copyright 2019 Kuei-chun Chen. All rights reserved.

package sim

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
	"time"

	"github.com/simagix/gox"
	"go.mongodb.org/mongo-driver/bson"
)

// PrintPerformanceStats prints performance stats, a place holder
func PrintPerformanceStats(filename string) string {
	var err error
	var data []byte
	var rd *bufio.Reader
	var metrics map[string][]bson.M
	var stats string

	if rd, err = gox.NewFileReader(filename); err != nil {
		return err.Error()
	}
	if data, err = ioutil.ReadAll(rd); err != nil {
		return err.Error()
	}
	buffer := bytes.NewBuffer(data)
	gob.Register(time.Duration(0))
	dec := gob.NewDecoder(buffer)
	if err = dec.Decode(&metrics); err != nil {
		return err.Error()
	}

	keys := []string{}
	for k := range metrics {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		tokens := strings.Split(k[1:], "-")
		stats += fmt.Sprintf("\nMinute: %v Connection %v:", tokens[0], tokens[1])
		durations := map[string][]time.Duration{}
		for _, res := range metrics[k] {
			for k, v := range res {
				durations[k] = append(durations[k], v.(time.Duration))
			}
		}
		keys := make([]string, 0, len(durations))
		for k := range durations {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := durations[k]
			sort.Slice(v, func(i int, j int) bool {
				return v[i] < v[j]
			})
			var sum time.Duration
			for _, t := range v {
				sum += t
			}
			length := len(v)
			p95 := int64(float64(length) * .95)
			p99 := int64(float64(length) * .99)
			stats += fmt.Sprintf("\n\t[%12s] (samples, min, avg, p95, p99, max) = (%v, %v, %v, %v, %v, %v)",
				k, length, v[0], sum/time.Duration(length), v[p95], v[p99], v[length-1])
		}
	}
	return stats
}
