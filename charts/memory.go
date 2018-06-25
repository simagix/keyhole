package charts

import (
	"encoding/json"
	"strconv"

	"github.com/simagix/keyhole/stats"
)

// GetMemoryCSV -
func GetMemoryCSV() []string {
	var docs []string
	docs = append(docs, "date,res,virt")
	for _, value := range stats.ChartsDocs {
		stat := stats.ServerStatusDoc{}
		for _, doc := range value {
			buf, _ := json.Marshal(doc)
			json.Unmarshal(buf, &stat)
			docs = append(docs, stat.LocalTime.Format("2006-01-02T15:04:05Z")+","+strconv.Itoa(stat.Mem.Resident)+","+strconv.Itoa(stat.Mem.Virtual))
		}
		break
	}

	return docs
}

// MemoryJS =
var MemoryJS = `
// set the dimensions and margins of the graph
var margin = {top: 20, right: 20, bottom: 30, left: 50},
    width = 640 - margin.left - margin.right,
    height = 320 - margin.top - margin.bottom;

// set the ranges
var x = d3.scaleTime().range([0, width]);
var y = d3.scaleLinear().range([height, 0]);
var z = d3.scaleOrdinal(d3.schemeCategory10);

// define the 1st line
var valueline = d3.line()
    .x(function(d) { return x(d.date); })
    .y(function(d) { return y(d.res); });

// define the 2nd line
var valueline2 = d3.line()
    .x(function(d) { return x(d.date); })
    .y(function(d) { return y(d.virt); });

// append the svg obgect to the body of the page
// appends a 'group' element to 'svg'
// moves the 'group' element to the top left margin
var svg = d3.select("body").append("svg")
    .attr("width", width + margin.left + margin.right)
    .attr("height", height + margin.top + margin.bottom)
  .append("g")
    .attr("transform",
          "translate(" + margin.left + "," + margin.top + ")");

// Get the data
d3.csv("v1/memory/csv", function(error, data) {
  if (error) throw error;

  // format the data
  data.forEach(function(d) {
      d.date = new Date(d.date)
      d.res = +d.res;
      d.virt = +d.virt;
  });

  // Scale the range of the data
  x.domain(d3.extent(data, function(d) { return d.date; }));
  y.domain([0, d3.max(data, function(d) {
	  return Math.max(d.res, d.virt); })]);

  // Add the valueline path.
  svg.append("path")
      .data([data])
      .attr("class", "line")
      .attr("d", valueline);

  // Add the valueline2 path.
  svg.append("path")
      .data([data])
      .attr("class", "line")
      .style("stroke", "red")
      .attr("d", valueline2);

  // Add the X Axis
  svg.append("g")
      .attr("transform", "translate(0," + height + ")")
      .call(d3.axisBottom(x));

  // Add the Y Axis
  svg.append("g")
      .call(d3.axisLeft(y));

});
`
