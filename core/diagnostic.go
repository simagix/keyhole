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
	"strings"

	"github.com/globalsign/mgo/bson"
)

// PrintDiagnosticData prints diagnostic data of MongoD
func PrintDiagnosticData(filename string, span int, isWeb bool) (string, error) {
	var err error
	var serverInfo interface{}
	var serverStatusList []ServerStatusDoc
	var replSetStatusList []ReplSetStatusDoc
	var fi os.FileInfo

	if fi, err = os.Stat(filename); err != nil {
		return "", err
	}
	switch mode := fi.Mode(); {
	case mode.IsDir():
		if serverInfo, serverStatusList, replSetStatusList, err = ReadDiagnosticDir(filename); err != nil {
			return "", err
		}
	case mode.IsRegular():
		if serverInfo, serverStatusList, replSetStatusList, err = AnalyzeServerStatus(filename); err != nil {
			log.Println(err)
			if serverInfo, serverStatusList, replSetStatusList, err = ReadDiagnosticFile(filename); err != nil {
				return "", err
			}
		}
	}

	if serverInfo != nil {
		b, _ := json.MarshalIndent(serverInfo, "", "  ")
		log.Println(string(b))
	}

	if span < 0 {
		span = int(serverStatusList[(len(serverStatusList)-1)].LocalTime.Sub(serverStatusList[0].LocalTime).Seconds()) / 20
	}

	if isWeb {
		var buf []byte
		var bmap = []bson.M{}
		var cmap = []bson.M{}
		buf, _ = json.Marshal(serverStatusList)
		json.Unmarshal(buf, &bmap)
		ChartsDocs["serverStatus"] = bmap
		buf, _ = json.Marshal(replSetStatusList)
		json.Unmarshal(buf, &cmap)
		ChartsDocs["replSetGetStatus"] = cmap
	}

	return PrintAllStats(serverStatusList, span), err
}

// ReadDiagnosticDir reads diagnotics.data from a directory
func ReadDiagnosticDir(dirname string) (interface{}, []ServerStatusDoc, []ReplSetStatusDoc, error) {
	var err error
	var serverInfo interface{}
	var serverStatusList []ServerStatusDoc
	var replSetStatusList []ReplSetStatusDoc
	var docs []ServerStatusDoc
	var repls []ReplSetStatusDoc
	var files []os.FileInfo

	if files, err = ioutil.ReadDir(dirname); err != nil {
		return serverInfo, serverStatusList, replSetStatusList, err
	}

	for _, f := range files {
		if strings.Index(f.Name(), "metrics.") != 0 && strings.Index(f.Name(), "keyhole_stats.") != 0 {
			continue
		}
		filename := dirname + "/" + f.Name()

		if serverInfo, docs, repls, err = AnalyzeServerStatus(filename); err != nil {
			if serverInfo, docs, repls, err = ReadDiagnosticFile(filename); err != nil {
				return serverInfo, serverStatusList, replSetStatusList, err
			}
			replSetStatusList = append(replSetStatusList, repls...)
		}

		serverStatusList = append(serverStatusList, docs...)
	}

	return serverInfo, serverStatusList, replSetStatusList, err
}

// ReadDiagnosticFile reads diagnostic.data from a file
func ReadDiagnosticFile(filename string) (interface{}, []ServerStatusDoc, []ReplSetStatusDoc, error) {
	var buffer []byte
	var err error
	var serverInfo interface{}
	var serverStatusList []ServerStatusDoc
	var replSetStatusList []ReplSetStatusDoc
	var docs []bson.M
	var repls []bson.M
	var pos int32

	if buffer, err = ioutil.ReadFile(filename); err != nil {
		return serverInfo, serverStatusList, replSetStatusList, err
	}

	log.Println("reading", filename)
	for {
		if pos >= int32(len(buffer)) {
			break
		}
		bs := buffer[pos:(pos + 4)]
		var length int32

		if err = binary.Read(bytes.NewReader(bs), binary.LittleEndian, &length); err != nil {
			return serverInfo, serverStatusList, replSetStatusList, err
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
				return serverInfo, serverStatusList, replSetStatusList, err
			}
			var bytesBufWriter bytes.Buffer
			writer := bufio.NewWriter(&bytesBufWriter)
			io.Copy(writer, r)
			r.Close()
			bson.Unmarshal(bytesBufWriter.Bytes(), &doc)
			// systemMetrics
			// end
			// start
			// serverStatus
			// replSetGetStatus
			// local.oplog.rs.stats

			if doc["serverStatus"] != nil {
				docs = append(docs, doc["serverStatus"].(bson.M))
			}

			if doc["replSetGetStatus"] != nil {
				repls = append(repls, doc["replSetGetStatus"].(bson.M))
			}
		} else {
			log.Println("==>", out["type"])
		}
	}

	if buffer, err = json.Marshal(docs); err != nil {
		return serverInfo, serverStatusList, replSetStatusList, err
	}
	json.Unmarshal(buffer, &serverStatusList)

	if buffer, err = json.Marshal(repls); err != nil {
		return serverInfo, serverStatusList, replSetStatusList, err
	}
	json.Unmarshal(buffer, &replSetStatusList)
	return serverInfo, serverStatusList, replSetStatusList, err
}
