// Copyright 2020 Kuei-chun Chen. All rights reserved.

package analytics

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// GetMetricsFilenames gets metrics or keyhole_stats filesnames
func GetMetricsFilenames(filenames []string) []string {
	var err error
	var fi os.FileInfo
	fnames := []string{}

	for _, filename := range filenames {
		if fi, err = os.Stat(filename); err != nil {
			continue
		}
		switch mode := fi.Mode(); {
		case mode.IsDir():
			files, _ := ioutil.ReadDir(filename)
			for _, file := range files {
				if file.IsDir() == false &&
					(strings.HasPrefix(file.Name(), "metrics.") || strings.HasPrefix(file.Name(), "keyhole_stats.")) {
					fnames = append(fnames, filename+"/"+file.Name())
				}
			}
		case mode.IsRegular():
			basename := filepath.Base(filename)
			if strings.HasPrefix(basename, "metrics.") || strings.HasPrefix(basename, "keyhole_stats.") {
				fnames = append(fnames, filename)
			}
		}
	}
	sort.Slice(fnames, func(i int, j int) bool {
		return fnames[i] < fnames[j]
	})
	return fnames
}

// GetScoreByRange gets score
func GetScoreByRange(v float64, low float64, high float64) int {
	if math.IsNaN(v) {
		return 101
	}
	var score int
	if v < low {
		score = 100
	} else if v > high {
		score = 0
	} else {
		score = int(100 * (1 - float64(v-low)/(high-low)))
	}
	return score
}

// GetShortLabel gets shorten label
func GetShortLabel(label string) string {
	if strings.HasPrefix(label, "conns_") {
		label = label[6:]
	} else if strings.HasPrefix(label, "cpu_") {
		label = label[4:]
	} else if strings.HasPrefix(label, "latency_") {
		label = label[8:]
	} else if strings.HasPrefix(label, "mem_") {
		label = label[4:]
	} else if strings.HasPrefix(label, "net_") {
		label = label[4:]
	} else if strings.HasPrefix(label, "ops_") {
		label = label[4:]
	} else if strings.HasPrefix(label, "q_active_") {
		label = label[9:]
	} else if strings.HasPrefix(label, "q_queued_") {
		label = label[9:]
	} else if strings.HasPrefix(label, "scan_") {
		label = label[5:]
	} else if strings.HasPrefix(label, "ticket_") {
		label = label[7:]
	} else if strings.HasPrefix(label, "wt_blkmgr_") {
		label = label[10:]
	} else if strings.HasPrefix(label, "wt_cache_") {
		label = label[9:]
	} else if strings.HasPrefix(label, "wt_dhandles_") {
		label = label[12:]
	} else if strings.HasPrefix(label, "wt_") {
		label = label[3:]
		if strings.HasSuffix(label, "_evicted") {
			label = label[:len(label)-8]
		}
	}
	return label
}

// GetFormulaHTML returns scoring formula
func GetFormulaHTML(metric string) string {
	html := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
	  <title>Metric Scores</title>
		<meta http-equiv="Cache-Control" content="no-cache, no-store, must-revalidate" />
		<meta http-equiv="Pragma" content="no-cache" />
		<meta http-equiv="Expires" content="0" />
	  <script src="https://www.gstatic.com/charts/loader.js"></script>
	  <style>
			body {
				font-family: Arial, Helvetica, sans-serif;
	      font-size; 11px;
			}
	    table
	    {
	      font-family: Consolas, monaco, monospace;
	      font-size; 10px;
	    	border-collapse:collapse;
	    	#width:100%;
	    	#max-width:800px;
	    	min-width:600px;
	    	#text-align:center;
	    }
	    caption
	    {
	    	caption-side:top;
	    	font-weight:bold;
	    	font-style:italic;
	    	margin:4px;
	    }
	    table,th, td
	    {
	    	border: 1px solid gray;
	    }
	    th, td
	    {
	    	height: 24px;
	    	padding:4px;
	    	vertical-align:middle;
	    }
	    th
	    {
	    	#background-image:url(table-shaded.png);
	      background-color: #333;
	      color: #FFF;
	    }
	    .rowtitle
	    {
	    	font-weight:bold;
	    }
			a {
			  text-decoration: none;
			  color: #000;
			  display: block;

			  -webkit-transition: font-size 0.3s ease, background-color 0.3s ease;
			  -moz-transition: font-size 0.3s ease, background-color 0.3s ease;
			  -o-transition: font-size 0.3s ease, background-color 0.3s ease;
			  -ms-transition: font-size 0.3s ease, background-color 0.3s ease;
			  transition: font-size 0.3s ease, background-color 0.3s ease;
			}
			a:hover {
				color: blue;
			}
	    .button {
	      background-color: #333;
	      border: none;
	      border-radius: 8px;
	      color: white;
	      padding: 3px 6px;
	      text-align: center;
	      text-decoration: none;
	      display: inline-block;
	      font-size: 12px;
	    }
	    </style>
	</head>
	<body><h3>Scores:</h3>
	<table>
	<tr><th>Metric</th><th>Formula</th><th>Low Watermark</th><th>High Watermark</th></tr>
	`
	p := message.NewPrinter(language.English)
	keys := []string{}
	for k := range FormulaMap {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i int, j int) bool {
		return keys[i] < keys[j]
	})
	for i := range keys {
		k := keys[i]
		value := FormulaMap[k]
		html += fmt.Sprintf(`<tr><td class='rowtitle'>%v</td><td>%v</td><td align='right'>%v</td><td align='right'>%v</td></tr>`,
			value.label, value.formula, p.Sprintf(`%v`, value.low), p.Sprintf(`%v`, value.high))
	}
	html += `</table></body></html>`
	return html
}
