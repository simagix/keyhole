// Copyright 2018 Kuei-chun Chen. All rights reserved.

package ftdc

import (
	"bytes"
	"errors"
	"io"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PathSeparator -
const PathSeparator = "/" // path separator

// Decode decodes MongoDB FTDC data
func (m *Metrics) decode(buffer []byte) (MetricsData, error) {
	var err error
	var dp = MetricsData{DataPointsMap: map[string][]int64{}, Buffer: buffer}

	r := bytes.NewReader(buffer)
	dp.DocSize = GetUint32(r) // first bson document length
	r.Seek(int64(dp.DocSize), io.SeekStart)
	dp.NumAttribs = GetUint32(r) // 4 bytes # of keys
	dp.NumDeltas = GetUint32(r)  // 4 bytes # of deltas
	ptr, _ := r.Seek(0, io.SeekCurrent)
	r = bytes.NewReader(buffer[ptr:]) // reset reader to where deltas begin

	// use DOM (bson.D) to ensure orders
	var attribsList = []string{}
	// systemMetrics
	// end
	// start
	// serverStatus
	// replSetGetStatus
	// local.oplog.rs.stats
	var docElem = bson.D{}
	bson.Unmarshal(buffer[:dp.DocSize], &docElem) // first document
	traverseDocElem(&attribsList, &dp.DataPointsMap, docElem, "")

	if len(dp.DataPointsMap) != int(dp.NumAttribs) || len(attribsList) != int(dp.NumAttribs) {
		return dp, errors.New("inconsistent FTDC data")
	}

	// deltas
	// d where d > 0, return d
	// 0d -> there are d number of zeros
	var delta int64
	var zerosLeft uint64
	for _, attr := range attribsList {
		list := dp.DataPointsMap[attr]
		v := list[0]
		for j := uint32(0); j < dp.NumDeltas; j++ {
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
		dp.DataPointsMap[attr] = list
	}
	return dp, err
}

func traverseDocElem(attribsList *[]string, attribsMap *map[string][]int64, docElem interface{}, parentPath string) {
	switch value := docElem.(type) {
	case bson.A:
		for i, v := range value {
			fld := parentPath + PathSeparator + strconv.Itoa(i)
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
			name := elem.Key
			if parentPath != "" {
				name = parentPath + PathSeparator + name
			}
			traverseDocElem(attribsList, attribsMap, elem.Value, name)
		}
	case primitive.Timestamp:
		tKey := parentPath + "/t"
		(*attribsMap)[tKey] = []int64{int64(0)}
		(*attribsList) = append((*attribsList), tKey)
		iKey := parentPath + "/i"
		(*attribsMap)[iKey] = []int64{int64(0)}
		(*attribsList) = append((*attribsList), iKey)
	case primitive.ObjectID: // ignore it
	case string: // ignore it
	case float64:
		(*attribsMap)[parentPath] = []int64{int64(value)}
		(*attribsList) = append((*attribsList), parentPath)
	case int:
		(*attribsMap)[parentPath] = []int64{int64(value)}
		(*attribsList) = append((*attribsList), parentPath)
	case int32:
		(*attribsMap)[parentPath] = []int64{int64(value)}
		(*attribsList) = append((*attribsList), parentPath)
	case int64:
		(*attribsMap)[parentPath] = []int64{value}
		(*attribsList) = append((*attribsList), parentPath)
	case primitive.DateTime: // ignore it
		(*attribsMap)[parentPath] = []int64{int64(value)}
		(*attribsList) = append((*attribsList), parentPath)
	default:
		// log.Fatalf("'%s' ==> %T\n", parentPath, value)
	}
}
