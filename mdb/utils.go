// Copyright 2020-present Kuei-chun Chen. All rights reserved.

package mdb

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	outdir = "./out"
)

// CodeDefault shows default color
var CodeDefault = "\x1b[0m"

// CodeRed shows red color
var CodeRed = "\x1b[31;1m"

// CodeGreen shows green color
var CodeGreen = "\x1b[32;1m"

// CodeYellow shows yellow color
var CodeYellow = "\x1b[33;1m"

// CodeBlue shows blue color
var CodeBlue = "\x1b[34;1m"

// ChartDataPoint has chart data points key/value
type ChartDataPoint struct {
	label string
	value int64
}

func getTopChartPoints(chartPoints []ChartDataPoint, numPoints int) (string, []ChartDataPoint) {
	unit := 1024 * 1024 * 1024
	for unit > 1 {
		count := 0
		points := []ChartDataPoint{}
		for _, v := range chartPoints {
			val := int64(float64(v.value) / float64(unit))
			if val > 1 {
				count++
			}
			points = append(points, ChartDataPoint{label: v.label, value: val})
		}
		if count >= numPoints/2 || count == len(chartPoints) {
			if unit == 1024*1024*1024 {
				return "GB", points
			} else if unit == 1024*1024 {
				return "MB", points
			} else if unit == 1024 {
				return "KB", points
			}
		}
		unit /= 1024
	}
	return "B", chartPoints
}

// ToInt converts to int
func ToInt(num interface{}) int {
	f := fmt.Sprintf("%v", num)
	x, _ := strconv.ParseFloat(f, 64)
	return int(x)
}

// ToInt32 converts to int32
func ToInt32(num interface{}) int32 {
	f := fmt.Sprintf("%v", num)
	x, _ := strconv.ParseFloat(f, 64)
	return int32(x)
}

// ToInt64 converts to in64
func ToInt64(num interface{}) int64 {
	return toInt64(num)
}
func toInt64(num interface{}) int64 {
	f := fmt.Sprintf("%v", num)
	x, _ := strconv.ParseFloat(f, 64)
	return int64(x)
}

// ToFloat64 converts to float64
func ToFloat64(num interface{}) float64 {
	return toFloat64(num)
}
func toFloat64(num interface{}) float64 {
	f := fmt.Sprintf("%v", num)
	x, _ := strconv.ParseFloat(f, 64)
	return x
}

// SplitNamespace returns db, collection
func SplitNamespace(namespace string) (string, string) {
	dot := strings.Index(namespace, ".")
	if dot < 0 {
		return namespace, ""
	}
	return namespace[:dot], namespace[dot+1:]
}

// SkipNamespace skips namespace
func SkipNamespace(namespace string, namespaceMap map[string]bool) bool {
	if len(namespaceMap) == 0 {
		return false
	}
	dbName, _ := SplitNamespace(namespace)
	allDB := dbName + ".*"
	if namespaceMap[allDB] || namespaceMap[namespace] {
		return false
	}
	return true
}

// DoesFileExist returns true if file exists
func DoesFileExist(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	if info == nil {
		return false
	}
	return !info.IsDir()
}
