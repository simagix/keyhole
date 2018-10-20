package keyhole

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"io"
	"io/ioutil"
	"log"

	"github.com/globalsign/mgo/bson"
)

// ReadDiagnosticDir reads diagnotics.data from a directory
func ReadDiagnosticDir(dirname string) {

}

// ReadDiagnosticFile reads diagnostic.data from a file
func ReadDiagnosticFile(filename string) (interface{}, []bson.M, error) {
	var buffer []byte
	var err error
	var serverInfo interface{}
	var serverStatusList []bson.M
	var pos int32

	if buffer, err = ioutil.ReadFile(filename); err != nil {
		return serverInfo, serverStatusList, err
	}

	for {
		if pos >= int32(len(buffer)) {
			break
		}
		bs := buffer[pos:(pos + 4)]
		var length int32

		if err = binary.Read(bytes.NewReader(bs), binary.LittleEndian, &length); err != nil {
			return serverInfo, serverStatusList, err
		}

		bs = buffer[pos:(pos + length)]
		var out = bson.M{}
		err = bson.Unmarshal(bs, &out)
		pos += length

		if err != nil {
			continue
		} else if out["type"] == 0 {
			serverInfo = out["doc"]
		} else if out["type"] == 1 {
			buf := out["data"].([]byte)
			var doc = bson.M{}
			// zlib decompress
			buf = buf[4:]
			bytesBuf := bytes.NewReader(buf)
			var r io.ReadCloser
			if r, err = zlib.NewReader(bytesBuf); err != nil {
				return serverInfo, serverStatusList, err
			}
			var bytesBufWriter bytes.Buffer
			writer := bufio.NewWriter(&bytesBufWriter)
			io.Copy(writer, r)
			r.Close()
			bson.Unmarshal(bytesBufWriter.Bytes(), &doc)
			if doc["serverStatus"] == nil {
				log.Println("Not serverStatus")
			} else {
				serverStatusList = append(serverStatusList, doc["serverStatus"].(bson.M))
			}
		}
	}

	return serverInfo, serverStatusList, err
}
