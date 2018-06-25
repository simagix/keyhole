package charts

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path[1:] == "memory" {
		str := strings.Replace(indexHTML, "__TITLE__", "Memory Stats Charts", -1)
		fmt.Fprintf(w, strings.Replace(str, "__MODULE__", "memory", -1))
	} else if r.URL.Path[1:] == "memory/index.js" {
		fmt.Fprintf(w, MemoryJS)
	} else if r.URL.Path[1:] == "v1/memory/csv" {
		fmt.Fprintf(w, strings.Join(GetMemoryCSV()[:], "\n"))
	} else {
		fmt.Fprintf(w, "Keyhole Performance Charts!  Unknow API!")
	}
}

// HTTPServer listens to port 5408
func HTTPServer(port int) {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}

var indexHTML = `
<!DOCTYPE html>
<meta charset="utf-8">
<meta http-equiv="refresh" content="30">
<style> /* set the CSS */

.line {
  fill: none;
  stroke: steelblue;
  stroke-width: 2px;
}

</style>
<body>
<h1>__TITLE__</h1>
<!-- load the d3.js library -->
<script src="https://d3js.org/d3.v4.min.js"></script>
<script src="__MODULE__/index.js"></script>
</body>
`
