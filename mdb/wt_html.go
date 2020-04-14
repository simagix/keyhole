// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

const html = `
<!DOCTYPE html>
<html lang="en">
<head>
<title>MongoDB WiredTiger Cache Usages</title>
<meta http-equiv="Cache-Control" content="no-cache, no-store, must-revalidate" />
<meta http-equiv="Pragma" content="no-cache" />
<meta http-equiv="Expires" content="0" />
<script src="https://www.gstatic.com/charts/loader.js"></script>
<style>
  body {
    font-family: Arial, Helvetica, sans-serif;
    font-size; 11px;
  }
</style>
</head>
<body>
<table>
<tr>
  <td><div id="topCaches"></div></td><td><div id="cacheDistr"></div></td>
</tr>
<tr>
  <td><div id="topDataCache"></div></td><td><div id="topIndexesCache"></div></td>
</tr>
</table>

<script>
  var chart_data;
	google.charts.load('current', {'packages':['corechart']});
	google.charts.setOnLoadCallback(drawCharts);

  function drawCharts(){
    redraw();
    var intervalID = setInterval(redraw, 5000);
  }

  function redraw() {
    var xmlhttp = new XMLHttpRequest();
    var url = "wt/data";
    xmlhttp.onreadystatechange = function() {
      if (this.readyState == 4 && this.status == 200) {
          var doc = JSON.parse(this.responseText);
          drawPieChart('topCaches', doc.top_caches.title, doc.top_caches.data)
          drawPieChart('cacheDistr', doc.cache_distr.title, doc.cache_distr.data)
          drawColumnChart('topDataCache', doc.top_data_cache.title, doc.top_data_cache.data)
          drawColumnChart('topIndexesCache', doc.top_indexes_cache.title, doc.top_indexes_cache.data)
      }
    };
    xmlhttp.open("GET", url, true);
    xmlhttp.send();
  }

  function drawPieChart(divID, title, data) {
    var chart_data = new google.visualization.arrayToDataTable(data);
    var options = {
      'title': title,
      'width': 600,
      'height': 360,
  		'is3D': true,
  		'titleTextStyle': {'fontSize': 20},
  		'chartArea': {'width': '95%'},
  		'legend': { 'position': 'labeled' }
  	};

    // Instantiate and draw our chart, passing in some options.
    var chart = new google.visualization.PieChart(document.getElementById(divID));
    chart.draw(chart_data, options);
  }

  function drawColumnChart(divID, title, data) {
    var chart_data = new google.visualization.arrayToDataTable(data);
    var options = {
      'title': title,
      'width': 600,
      'height': 360,
  		'titleTextStyle': {'fontSize': 20},
  		'legend': { 'position': 'labeled' }
  	};

    // Instantiate and draw our chart, passing in some options.
    var chart = new google.visualization.ColumnChart(document.getElementById(divID));
    chart.draw(chart_data, options);
  }
</script>
</body>
`
