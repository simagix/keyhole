package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/simagix/keyhole/stats"
	"gopkg.in/mgo.v2/bson"
)

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path[1:] == "index.js" {
		str := strings.Replace(indexJS, "__API__", "/v1/mem/resident/servers", 1)
		fmt.Fprintf(w, str)
	} else if r.URL.Path[1:] == "index.css" {
		fmt.Fprintf(w, indexCSS)
	} else if r.URL.Path[1:] == "v1/mem/resident/servers" {
		for _, value := range stats.ChartsDocs {
			stat := stats.ServerStatusDoc{}
			jsons := bson.M{}
			for _, doc := range value {
				buf, _ := json.Marshal(doc)
				json.Unmarshal(buf, &stat)
				jsons[stat.LocalTime.Format("2006-01-02T15:04:05Z")] = stat.Mem.Resident
			}

			jsonDoc := bson.M{"pts": jsons}
			buf, _ := json.Marshal(jsonDoc)
			fmt.Fprintf(w, string(buf))
		}
	} else {
		fmt.Fprintf(w, strings.Replace(indexHTML, "__TITLE__", "Resident Memory", 1))
	}
}

// HTTPServer listens to port 5408
func HTTPServer(port int) {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}

var indexHTML = `
<html>
    <head>
        <link rel="stylesheet" href="index.css">
        <title>Learn D3.js</title>
    </head>
    <body>
        <h1>__TITLE__</h1>

        <svg class="line-chart"></svg>

        <script src="https://d3js.org/d3.v4.min.js"></script>
        <script src="index.js"></script>
    </body>
</html>
`

var indexJS = `
//API to fetch data
// const api = 'https://api.coindesk.com/v1/bpi/historical/close.json?start=2017-12-31&end=2018-04-01';
const api = '__API__';
console.log(api);
/**
 * Loading data from API when DOM Content has been loaded'.
 */
document.addEventListener("DOMContentLoaded", function(event) {
setInterval(function(){
  fetch(api)
      .then(function(response) { return response.json(); })
      .then(function(data) {
          var parsedData = parseData(data);
          drawChart(parsedData);
      })
      .catch(function(err) { console.log(err); })
  });
}, 30*000);

/**
 * Parse data into key-value pairs
 * @param {object} data Object containing data points
 */
function parseData(data) {
    var arr = [];
    for (var i in data.pts) {
        arr.push({
            date: new Date(i), //date
            value: +data.pts[i] //convert string to number
        });
    }
    return arr;
}

/**
 * Creates a chart using D3
 * @param {object} data Object containing historical data of BPI
 */
function drawChart(data) {
d3.selectAll("g > *").remove()
var svgWidth = 600, svgHeight = 400;
var margin = { top: 20, right: 20, bottom: 30, left: 50 };
var width = svgWidth - margin.left - margin.right;
var height = svgHeight - margin.top - margin.bottom;

var svg = d3.select('svg')
    .attr("width", svgWidth)
    .attr("height", svgHeight);

var g = svg.append("g")
    .attr("transform", "translate(" + margin.left + "," + margin.top + ")");

var x = d3.scaleTime()
    .rangeRound([0, width]);

var y = d3.scaleLinear()
    .rangeRound([height, 0]);

var line = d3.line()
    .x(function(d) { return x(d.date)})
    .y(function(d) { return y(d.value)})
    x.domain(d3.extent(data, function(d) { return d.date }));
    y.domain(d3.extent(data, function(d) { return d.value }));

g.append("g")
    .attr("transform", "translate(0," + height + ")")
    .call(d3.axisBottom(x))
    .select(".domain")
    .remove();

g.append("g")
    .call(d3.axisLeft(y))
    .append("text")
    .attr("fill", "#000")
    .attr("transform", "rotate(-90)")
    .attr("y", 6)
    .attr("dy", "0.71em")
    .attr("text-anchor", "end")
    .text("Count");

g.append("path")
    .datum(data)
    .attr("fill", "none")
    .attr("stroke", "steelblue")
    .attr("stroke-linejoin", "round")
    .attr("stroke-linecap", "round")
    .attr("stroke-width", 1.5)
    .attr("d", line);
}
`

var indexCSS = `
html, body {
    margin: 0;
    padding: 0;
}
`
