package keyhole

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/globalsign/mgo/bson"
)

// PrintDiagnosticData prints diagnostic data of MongoD
func PrintDiagnosticData(filename string, span int) error {
	var err error
	var serverInfo interface{}
	var serverStatusList []bson.M
	var docs []ServerStatusDoc
	var fi os.FileInfo

	if fi, err = os.Stat(filename); err != nil {
		return err
	}
	switch mode := fi.Mode(); {
	case mode.IsDir():
		if serverInfo, serverStatusList, err = ReadDiagnosticDir(filename); err != nil {
			return err
		}
	case mode.IsRegular():
		if err = AnalyzeServerStatus(filename, span, false); err != nil {
			if serverInfo, serverStatusList, err = ReadDiagnosticFile(filename); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	for _, ss := range serverStatusList {
		b, _ := json.Marshal(ss)
		doc := ServerStatusDoc{}
		json.Unmarshal(b, &doc)
		docs = append(docs, doc)
	}

	if serverInfo != nil {
		b, _ := json.MarshalIndent(serverInfo, "", "  ")
		log.Println(string(b))
	}
	PrintAllStats(docs, span)
	return err
}

// ReadDiagnosticDir reads diagnotics.data from a directory
func ReadDiagnosticDir(dirname string) (interface{}, []bson.M, error) {
	var err error
	var serverInfo interface{}
	var serverStatusList []bson.M
	var info interface{}
	var docs []bson.M

	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		filename := dirname + "/" + f.Name()
		if info, docs, err = ReadDiagnosticFile(filename); err != nil {
			return serverInfo, serverStatusList, err
		}
		if info != nil {
			serverInfo = info
		}
		serverStatusList = append(serverStatusList, docs...)
	}
	return serverInfo, serverStatusList, err
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
