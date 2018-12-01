// Copyright 2018 Kuei-chun Chen. All rights reserved.

package ftdc

import (
	"bytes"
	"errors"
	"io"
	"log"
	"strconv"
	"time"

	"github.com/globalsign/mgo/bson"
)

// PathSeparator -
const PathSeparator = "/" // path separator

// Decode decodes MongoDB FTDC data
func (f *FTDC) Decode(data []byte) error {
	var err error

	r := bytes.NewReader(data)
	size := GetUint32(r) // first bson document length
	r.Seek(int64(size), io.SeekStart)
	f.NumAttribs = GetUint32(r) // 4 bytes # of keys
	f.NumDeltas = GetUint32(r)  // 4 bytes # of deltas
	ptr, _ := r.Seek(0, io.SeekCurrent)
	r = bytes.NewReader(data[ptr:]) // reset reader to where deltas begin

	// use DOM (bson.D) to ensure orders
	var attribsList = []string{}
	var docElem = bson.D{}
	bson.Unmarshal(data, &docElem) // first document
	traverseDocElem(&attribsList, &f.DataPointsMap, docElem, "")

	if len(f.DataPointsMap) != int(f.NumAttribs) || len(attribsList) != int(f.NumAttribs) {
		return errors.New("inconsistent FTDC data")
	}

	// deltas
	// d where d > 0, return d
	// 0d -> there are d number of zeros
	var delta int64
	var zerosLeft uint64
	for _, attr := range attribsList {
		list := f.DataPointsMap[attr]
		v := list[0]
		for j := uint32(0); j < f.NumDeltas; j++ {
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
		f.DataPointsMap[attr] = list
	}
	return err
}

func traverseDocElem(attribsList *[]string, attribsMap *map[string][]int64, docElem interface{}, parentPath string) {
	switch value := docElem.(type) {
	case []interface{}:
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
			name := elem.Name
			if parentPath != "" {
				name = parentPath + PathSeparator + name
			}
			traverseDocElem(attribsList, attribsMap, elem.Value, name)
		}
	case bson.MongoTimestamp:
		tKey := parentPath + "/t"
		iKey := parentPath + "/i"
		(*attribsMap)[tKey] = []int64{int64(value)}
		(*attribsList) = append((*attribsList), tKey)
		(*attribsMap)[iKey] = []int64{int64(value)}
		(*attribsList) = append((*attribsList), iKey)
	case bson.ObjectId, string: // ignore it
		// (*attribsMap)[parentPath] = []int{}
	case float64:
		(*attribsMap)[parentPath] = []int64{int64(value)}
		(*attribsList) = append((*attribsList), parentPath)
	case int:
		(*attribsMap)[parentPath] = []int64{int64(value)}
		(*attribsList) = append((*attribsList), parentPath)
	case int64:
		(*attribsMap)[parentPath] = []int64{value}
		(*attribsList) = append((*attribsList), parentPath)
	case time.Time:
		(*attribsMap)[parentPath] = []int64{time.Time(value).UnixNano() / 1000000}
		(*attribsList) = append((*attribsList), parentPath)
	default:
		log.Fatalf("%s %T\n", parentPath, value)
	}
}
