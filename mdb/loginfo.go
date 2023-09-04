// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/simagix/gox"
)

// COLLSCAN constance
const COLLSCAN = "COLLSCAN"

// LogInfo keeps loginfo struct
type LogInfo struct {
	Collscan   bool        `bson:"collscan"`
	DBVersion  string      `bson:"version"`
	Histograms []Histogram `bson:"histogram"`
	Logger     *gox.Logger `bson:"keyhole"`
	LogType    string      `bson:"type"`
	Regex      string      `bson:"regex"`
	OpPatterns []OpPattern `bson:"opPatterns"`
	Redaction  bool        `bson:"redact"`
	SlowOps    []RawLog    `bson:"slowOps"`

	filename string
	logs     []string
	regex    string
	regexp   *regexp.Regexp
	silent   bool
	verbose  bool
}

// OpPattern stores performance data
type OpPattern struct {
	Command     string `bson:"command"`     // count, delete, find, remove, and update
	Count       int    `bson:"count"`       // number of ops
	Filter      string `bson:"filter"`      // query pattern
	MaxMilli    int    `bson:"maxmilli"`    // max millisecond
	Namespace   string `bson:"ns"`          // database.collectin
	Scan        string `bson:"scan"`        // COLLSCAN
	TotalMilli  int64  `bson:"totalmilli"`  // total milliseconds
	TotalReslen int64  `bson:"totalreslen"` // total reslen
	Index       string `bson:"index"`       // index used
}

// RawLog holds slow ops log and time
type RawLog struct {
	Milli int
	Log   string
}

// LogStats log stats structure
type LogStats struct {
	filter string
	index  string
	milli  int
	ns     string
	op     string
	reslen int
	scan   string
	utc    string
}

// Histogram stores ops info
type Histogram struct {
	UTC string         `bson:"utc"`
	Ops map[string]int `bson:"Ops"`
}

const dollarCmd = "$cmd"

// NewLogInfo -
func NewLogInfo(version string) *LogInfo {
	li := LogInfo{Logger: gox.GetLogger(version), Collscan: false, silent: false, verbose: false}
	// li.regex = `^(\S+) \S+\s+(\w+)\s+\[\w+\] (\w+) (\S+) \S+: (.*) (\d+)ms$` // SERVER-37743
	li.regex = `^(\S+) \S+\s+(\w+)\s+\[\w+\] (warning: log .* \.\.\. )?(\w+) (\S+) \S+: (.*) (\d+)ms$` // SERVER-37743
	return &li
}

// SetCollscan -
func (li *LogInfo) SetCollscan(collscan bool) {
	li.Collscan = collscan
}

// SetRedaction sets redaction
func (li *LogInfo) SetRedaction(redaction bool) {
	li.Redaction = redaction
}

// SetSilent -
func (li *LogInfo) SetSilent(silent bool) {
	li.silent = silent
}

// SetVerbose -
func (li *LogInfo) SetVerbose(verbose bool) {
	li.verbose = verbose
}

// SetRegexPattern sets Regex patthen
func (li *LogInfo) SetRegexPattern(regex string) {
	if regex != "" {
		li.regex = regex
	}
}

const (
	logExt = "-log.bson.gz"
	topN   = 25
)

// AnalyzeFile analyze logs from a file
func (li *LogInfo) AnalyzeFile(filename string) error {
	var err error
	li.filename = filename
	if strings.HasSuffix(filename, logExt) {
		var data []byte
		var err error
		var fd *bufio.Reader
		if fd, err = gox.NewFileReader(filename); err != nil {
			return err
		}
		if data, err = io.ReadAll(fd); err != nil {
			return err
		}
		if err = bson.Unmarshal(data, &li); err != nil {
			return err
		}
	} else {
		li.LogType = ""
		var file *os.File
		var reader *bufio.Reader
		li.filename = filename
		if file, err = os.Open(filename); err != nil {
			return err
		}
		defer file.Close()
		if reader, err = gox.NewReader(file); err != nil {
			return err
		}
		var lineCounts int
		if !li.silent {
			lineCounts, _ = gox.CountLines(reader)
		}
		file.Seek(0, 0)
		if reader, err = gox.NewReader(file); err != nil {
			return err
		}
		if err = li.Parse(reader, lineCounts); err != nil {
			return err
		}
	}
	return nil
}

// Parse parse text or json
func (li *LogInfo) Parse(reader *bufio.Reader, counts ...int) error {
	var err error
	var buf []byte
	var isPrefix bool
	var opsMap map[string]OpPattern
	var stat LogStats
	opsMap = make(map[string]OpPattern)
	lineCounts := 0
	if len(counts) > 0 {
		lineCounts = counts[0]
	}
	index := 0
	var ts string
	var hist = Histogram{Ops: map[string]int{}}
	li.regexp = regexp.MustCompile(li.regex)
	for {
		if lineCounts > 0 && !li.silent && index%50 == 0 {
			fmt.Fprintf(os.Stderr, "\r%3d%% \r", (100*index)/lineCounts)
		}
		if buf, isPrefix, err = reader.ReadLine(); err != nil { // 0x0A separator = newline
			break
		}
		index++
		if len(buf) == 0 {
			continue
		}
		str := string(buf)
		for isPrefix {
			var bbuf []byte
			if bbuf, isPrefix, err = reader.ReadLine(); err != nil {
				break
			}
			str += string(bbuf)
		}
		if li.LogType == "" { //examine the log logType
			if regexp.MustCompile("^{.*}$").MatchString(str) {
				li.LogType = "logv2"
				if stat, err = li.ParseLogv2(str); err != nil {
					continue
				}
			} else {
				li.LogType = "text"
				if stat, err = li.ParseLog(str); err != nil {
					continue
				}
			}
		} else if li.LogType == "logv2" || str[0:1] == "{" {
			if li.DBVersion == "" && strings.Index(str, `"Build Info"`) > 0 {
				label := `"version":"`
				idx := strings.Index(str, label)
				if idx > 0 {
					li.DBVersion = str[idx+len(label):]
					idx = strings.Index(li.DBVersion, `"`)
					li.DBVersion = "v" + li.DBVersion[:idx]
				}
			} else if stat, err = li.ParseLogv2(str); err != nil {
				continue
			}
			li.LogType = "logv2"
		} else {
			if li.DBVersion == "" && strings.Index(str, " CONTROL ") > 0 {
				label := "db version"
				idx := strings.Index(str, label)
				if idx > 0 {
					li.DBVersion = str[idx+len(label)+1:]
				}
			} else if stat, err = li.ParseLog(str); err != nil {
				continue
			}
		}
		if stat.op == "" {
			if li.verbose {
				fmt.Println(str)
			}
			continue
		} else if stat.op == dollarCmd {
			continue
		}
		if stat.utc != ts { //	push
			fmt.Fprintf(os.Stderr, "\r     %v\r", stat.utc)
			if ts != "" {
				li.Histograms = append(li.Histograms, hist)
			}
			ts = stat.utc
			hist = Histogram{UTC: ts, Ops: map[string]int{}}
		}
		cnt := hist.Ops[stat.op]
		cnt++
		hist.Ops[stat.op] = cnt
		key := stat.op + "." + stat.ns + "." + stat.filter + "." + stat.scan
		_, ok := opsMap[key]
		if stat.op != "insert" && (len(li.SlowOps) < topN || stat.milli > li.SlowOps[topN-1].Milli) {
			li.SlowOps = append(li.SlowOps, RawLog{Milli: stat.milli, Log: str})
			sort.Slice(li.SlowOps, func(i, j int) bool {
				return li.SlowOps[i].Milli > li.SlowOps[j].Milli
			})
			if len(li.SlowOps) > topN {
				li.SlowOps = li.SlowOps[:topN]
			}
		}

		if ok {
			max := opsMap[key].MaxMilli
			if stat.milli > max {
				max = stat.milli
			}
			x := opsMap[key].TotalMilli + int64(stat.milli)
			y := opsMap[key].Count + 1
			z := opsMap[key].TotalReslen + int64(stat.reslen)
			opsMap[key] = OpPattern{Command: opsMap[key].Command, Namespace: stat.ns, Filter: opsMap[key].Filter,
				MaxMilli: max, TotalMilli: x, Count: y, Scan: stat.scan, Index: stat.index, TotalReslen: z}
		} else {
			opsMap[key] = OpPattern{Command: stat.op, Namespace: stat.ns, Filter: stat.filter, TotalMilli: int64(stat.milli),
				MaxMilli: stat.milli, Count: 1, Scan: stat.scan, Index: stat.index, TotalReslen: int64(stat.reslen)}
			li.logs = append(li.logs, str) // append a sample
		}
	}
	li.Histograms = append(li.Histograms, hist)
	li.OpPatterns = make([]OpPattern, 0, len(opsMap))
	for _, value := range opsMap {
		li.OpPatterns = append(li.OpPatterns, value)
	}
	sort.Slice(li.OpPatterns, func(i, j int) bool {
		return float64(li.OpPatterns[i].TotalMilli)/float64(li.OpPatterns[i].Count) > float64(li.OpPatterns[j].TotalMilli)/float64(li.OpPatterns[j].Count)
	})
	if !li.silent {
		fmt.Fprintf(os.Stderr, "\r                         \r")
	}
	return nil
}

// OutputBSON writes loginfo bson data
func (li *LogInfo) OutputBSON() (string, []byte, error) {
	var err error
	var data []byte
	var ofile string
	if len(li.OpPatterns) == 0 {
		return "", data, err
	}
	ofile = filepath.Base(li.filename)
	var bsonf, tsvf string
	ofile = strings.TrimSuffix(ofile, ".gz")
	if !strings.HasSuffix(ofile, ".log") {
		bsonf += ofile + logExt
		tsvf += ofile + ".tsv"
	} else {
		bsonf = ofile[:len(ofile)-4] + logExt
		tsvf = ofile[:len(ofile)-4] + ".tsv"
	}
	if li.Redaction {
		li.SlowOps = []RawLog{}
	}
	if li.LogType == "text" {
		li.Regex = li.regex
	} else if li.LogType == "logv2" {
		li.Regex = ""
	}

	var buffer bytes.Buffer
	if data, err = bson.Marshal(li); err != nil {
		return ofile, data, err
	}
	nw := 0
	var n int
	for nw < len(data) {
		if n, err = buffer.Write(data); err != nil {
			return ofile, data, err
		}
		nw += n
	}

	if !li.Redaction {
		for _, log := range li.logs {
			if data, err = bson.Marshal(bson.M{"raw": log}); err != nil {
				continue
			}
			if _, err = buffer.Write(data); err != nil {
				return ofile, data, err
			}
		}
	}

	os.Mkdir(outdir, 0755)
	idx := strings.Index(bsonf, logExt)
	basename := bsonf[:idx]
	ofile = fmt.Sprintf(`%v/%v%v`, outdir, basename, logExt)
	i := 1
	for DoesFileExist(ofile) {
		ofile = fmt.Sprintf(`%v/%v.%d%v`, outdir, basename, i, logExt)
		i++
	}
	if err = gox.OutputGzipped(buffer.Bytes(), ofile); err != nil {
		fmt.Println("write error:", err)
	} else {
		fmt.Println("bson log info written to", ofile)
	}

	// output TSV file
	re := regexp.MustCompile(`\r?\n`)
	lines := []string{fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v", "Row", "Category", "Avg Time", "Max Time", "Count", "Total Time", "Total Reslen", "Namespace", "COLLSCAN", "Index(es) Used", "Query Pattern")}
	for i, doc := range li.OpPatterns {
		avg := float64(doc.TotalMilli) / float64(doc.Count)
		lines = append(lines, fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v", i+1, doc.Command, gox.MilliToTimeString(avg), doc.MaxMilli, doc.Count,
			doc.TotalMilli, doc.TotalReslen, doc.Namespace, doc.Scan, re.ReplaceAllString(doc.Index, " "), doc.Filter))
	}

	idx = strings.Index(tsvf, ".tsv")
	basename = tsvf[:idx]
	tsv := fmt.Sprintf(`%v/%v.tsv`, outdir, basename)
	i = 1
	for DoesFileExist(tsv) {
		tsv = fmt.Sprintf(`%v/%v.%d.tsv`, outdir, basename, i)
		i++
	}

	tsvData := []byte(strings.Join(lines, "\n"))
	if err = os.WriteFile(tsv, tsvData, 0644); err != nil {
		fmt.Println("write error:", err)
	} else {
		fmt.Println("TSV log info written to", tsv)
	}
	return ofile, buffer.Bytes(), err
}

// OutputJSON writes json data to a file
func (li *LogInfo) OutputJSON() error {
	var err error
	var data []byte
	if data, err = bson.MarshalExtJSON(li, false, false); err != nil {
		return err
	}
	os.Mkdir(outdir, 0755)
	ofile := fmt.Sprintf("%v/%v", outdir, strings.ReplaceAll(filepath.Base(li.filename), "bson.gz", "json"))
	os.WriteFile(ofile, data, 0644)
	fmt.Println("json data written to", ofile)
	return err
}

// Print prints indexes
func (li *LogInfo) Print() {
	fmt.Println(li.printLogsSummary())
}

// printLogsSummary prints loginfo summary
func (li *LogInfo) printLogsSummary() string {
	var maxSize = 10
	red := CodeRed
	green := CodeGreen
	tail := CodeDefault
	if li.silent {
		red = ""
		green = ""
		tail = ""
	}
	summaries := []string{}
	var buffer bytes.Buffer
	buffer.WriteString("\r+----------+--------+------+--------+------+---------------------------------+--------------------------------------------------------------+\n")
	buffer.WriteString(fmt.Sprintf("| Command  |COLLSCAN|avg ms| max ms | Count| %-32s| %-60s |\n", "Namespace", "Query Pattern"))
	buffer.WriteString("|----------+--------+------+--------+------+---------------------------------+--------------------------------------------------------------|\n")
	count := 0
	for _, value := range li.OpPatterns {
		count++
		if count >= maxSize {
			break
		}
		str := value.Filter
		if len(value.Command) > 10 {
			value.Command = value.Command[:10]
		}
		if len(value.Namespace) > 33 {
			length := len(value.Namespace)
			value.Namespace = value.Namespace[:1] + "*" + value.Namespace[(length-31):]
		}
		if len(str) > 60 {
			str = value.Filter[:60]
			idx := strings.LastIndex(str, " ")
			if idx > 0 {
				str = value.Filter[:idx]
			}
		}
		output := ""
		avg := float64(value.TotalMilli) / float64(value.Count)
		avgstr := gox.MilliToTimeString(avg)
		if value.Scan == COLLSCAN {
			output = fmt.Sprintf("|%-10s %v%8s%v %6s %8d %6d %-33s %v%-62s%v|\n", value.Command, red, value.Scan, tail,
				avgstr, value.MaxMilli, value.Count, value.Namespace, red, str, tail)
		} else {
			output = fmt.Sprintf("|%-10s %8s %6s %8d %6d %-33s %-62s|\n", value.Command, value.Scan,
				avgstr, value.MaxMilli, value.Count, value.Namespace, str)
		}
		buffer.WriteString(output)
		if len(value.Filter) > 60 {
			remaining := value.Filter[len(str):]
			for i := 0; i < len(remaining); i += 60 {
				epos := i + 60
				var pstr string
				if epos > len(remaining) {
					epos = len(remaining)
					pstr = remaining[i:epos]
				} else {
					str = strings.Trim(remaining[i:epos], " ")
					idx := strings.LastIndex(str, " ")
					if idx >= 0 {
						pstr = str[:idx]
						i -= (60 - idx)
					} else {
						pstr = str
						i -= (60 - len(str))
					}
				}
				if value.Scan == COLLSCAN {
					output = fmt.Sprintf("|%74s   %v%-62s%v|\n", " ", red, pstr, tail)
					buffer.WriteString(output)
				} else {
					output = fmt.Sprintf("|%74s   %-62s|\n", " ", pstr)
					buffer.WriteString(output)
				}
			}
		}
		if value.Index != "" {
			output = fmt.Sprintf("|...index:  %v%-128s%v|\n", green, value.Index, tail)
			buffer.WriteString(output)
		}
	}
	buffer.WriteString("+----------+--------+------+--------+------+---------------------------------+--------------------------------------------------------------+\n")
	summaries = append(summaries, buffer.String())
	summaries = append(summaries, fmt.Sprintf(`top %d of %v lines displayed; see HTML report for details.`, count, len(li.OpPatterns)))
	return strings.Join(summaries, "\n")
}
