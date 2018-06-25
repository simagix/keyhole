package charts

import (
	"encoding/json"
	"strconv"

	"github.com/simagix/keyhole/stats"
)

// GetMemoryTSV -
func GetMemoryTSV() []string {
	var docs []string
	docs = append(docs, "date\tResident\tVirtual")
	for _, value := range stats.ChartsDocs {
		stat := stats.ServerStatusDoc{}
		for _, doc := range value {
			buf, _ := json.Marshal(doc)
			json.Unmarshal(buf, &stat)
			docs = append(docs, stat.LocalTime.Format("2006-01-02T15:04:05Z")+"\t"+strconv.Itoa(stat.Mem.Resident)+"\t"+strconv.Itoa(stat.Mem.Virtual))
		}
		break
	}

	return docs
}

// MemoryJS =
var MemoryJS = `
var svg = d3.select("svg"),
    margin = {top: 20, right: 80, bottom: 30, left: 50},
    width = svg.attr("width") - margin.left - margin.right,
    height = svg.attr("height") - margin.top - margin.bottom,
    g = svg.append("g").attr("transform", "translate(" + margin.left + "," + margin.top + ")");

var x = d3.scaleTime().range([0, width]),
    y = d3.scaleLinear().range([height, 0]),
    z = d3.scaleOrdinal(d3.schemeCategory10);

var line = d3.line()
    .curve(d3.curveBasis)
    .x(function(d) { return x(d.date); })
    .y(function(d) { return y(d.memory); });

// Get the data
d3.tsv("v1/memory/tsv", type, function(error, data) {
	if (error) throw error;

  var mems = data.columns.slice(1).map(function(id) {
    return {
      id: id,
      values: data.map(function(d) {
        return {date: d.date, memory: d[id]};
      })
    };
  });

  x.domain(d3.extent(data, function(d) { return d.date; }));

  y.domain([
    d3.min(mems, function(c) { return d3.min(c.values, function(d) { return d.memory; }); }),
    d3.max(mems, function(c) { return d3.max(c.values, function(d) { return d.memory; }); })
  ]);

  z.domain(mems.map(function(c) { return c.id; }));

  g.append("g")
      .attr("class", "axis axis--x")
      .attr("transform", "translate(0," + height + ")")
      .call(d3.axisBottom(x));

  g.append("g")
      .attr("class", "axis axis--y")
      .call(d3.axisLeft(y))
    .append("text")
      .attr("transform", "rotate(-90)")
      .attr("y", 6)
      .attr("dy", "0.71em")
      .attr("fill", "#000")
      .text("Memory, MB");

  var city = g.selectAll(".city")
    .data(mems)
    .enter().append("g")
      .attr("class", "city");

  city.append("path")
      .attr("class", "line")
      .attr("d", function(d) { return line(d.values); })
      .style("stroke", function(d) { return z(d.id); });

  city.append("text")
      .datum(function(d) { return {id: d.id, value: d.values[d.values.length - 1]}; })
      .attr("transform", function(d) { return "translate(" + x(d.value.date) + "," + y(d.value.memory) + ")"; })
      .attr("x", 3)
      .attr("dy", "0.35em")
      .style("font", "10px sans-serif")
      .text(function(d) { return d.id; });
});

function type(d, _, columns) {
  d.date = new Date(d.date);
  for (var i = 1, n = columns.length, c; i < n; ++i) d[c = columns[i]] = +d[c];
  return d;
}
`
