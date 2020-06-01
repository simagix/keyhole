// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"fmt"
	"reflect"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var codeRed = "\x1b[31;1m"
var codeGreen = "\x1b[32;1m"
var codeYellow = "\x1b[33;1m"
var codeBlue = "\x1b[34;1m"

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

func toArray(array interface{}) []bson.M {
	var mapArray []bson.M
	if array == nil {
		return mapArray
	}
	if reflect.TypeOf(array).String() == "primitive.A" {
		for _, t := range array.(primitive.A) {
			mapArray = append(mapArray, t.(bson.M))
		}
	} else {
		mapArray = array.([]bson.M)
	}
	return mapArray
}

func toInt64(num interface{}) int64 {
	f := fmt.Sprintf("%v", num)
	x, _ := strconv.ParseFloat(f, 64)
	return int64(x)
}
