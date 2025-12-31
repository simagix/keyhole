// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

const html = `
<!DOCTYPE html>
<html lang="en">
<head>
<title>MongoDB WiredTiger Cache Monitor</title>
<meta http-equiv="Cache-Control" content="no-cache, no-store, must-revalidate" />
<meta http-equiv="Pragma" content="no-cache" />
<meta http-equiv="Expires" content="0" />
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<script src="https://www.gstatic.com/charts/loader.js"></script>
<style>
  :root {
    --bg-primary: #0d1117;
    --bg-secondary: #161b22;
    --bg-card: #21262d;
    --border-color: #30363d;
    --text-primary: #e6edf3;
    --text-secondary: #8b949e;
    --accent-green: #3fb950;
    --accent-blue: #58a6ff;
    --accent-purple: #a371f7;
    --accent-orange: #d29922;
  }

  * {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
  }

  body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Noto Sans', Helvetica, Arial, sans-serif;
    background: var(--bg-primary);
    color: var(--text-primary);
    min-height: 100vh;
    padding: 24px;
  }

  .header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 24px;
    padding-bottom: 16px;
    border-bottom: 1px solid var(--border-color);
  }

  .header h1 {
    font-size: 24px;
    font-weight: 600;
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .header h1::before {
    content: '🍃';
    font-size: 28px;
  }

  .status {
    display: flex;
    align-items: center;
    gap: 8px;
    color: var(--text-secondary);
    font-size: 14px;
  }

  .status-dot {
    width: 8px;
    height: 8px;
    background: var(--accent-green);
    border-radius: 50%;
    animation: pulse 2s infinite;
  }

  .status-dot.loading {
    background: var(--accent-orange);
  }

  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.5; }
  }

  .dashboard {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 20px;
    max-width: 1400px;
    margin: 0 auto;
  }

  @media (max-width: 1024px) {
    .dashboard {
      grid-template-columns: 1fr;
    }
  }

  .card {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 12px;
    padding: 20px;
    transition: border-color 0.2s ease;
  }

  .card:hover {
    border-color: var(--accent-blue);
  }

  .card-title {
    font-size: 14px;
    font-weight: 500;
    color: var(--text-secondary);
    margin-bottom: 16px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .chart-container {
    display: flex;
    justify-content: center;
    align-items: center;
    min-height: 320px;
  }

  .loading-placeholder {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 16px;
    color: var(--text-secondary);
    min-height: 320px;
  }

  .spinner {
    width: 40px;
    height: 40px;
    border: 3px solid var(--border-color);
    border-top-color: var(--accent-blue);
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .loading-text {
    font-size: 14px;
  }

  .no-data {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 12px;
    color: var(--text-secondary);
    min-height: 320px;
  }

  .no-data-icon {
    font-size: 48px;
    opacity: 0.5;
  }

  .footer {
    text-align: center;
    margin-top: 24px;
    padding-top: 16px;
    border-top: 1px solid var(--border-color);
    color: var(--text-secondary);
    font-size: 12px;
  }

  .footer a {
    color: var(--accent-blue);
    text-decoration: none;
  }

  .footer a:hover {
    text-decoration: underline;
  }
</style>
</head>
<body>

<div class="header">
  <h1>WiredTiger Cache Monitor</h1>
  <div class="status">
    <span class="status-dot loading" id="statusDot"></span>
    <span id="lastUpdate">Collecting data...</span>
  </div>
</div>

<div class="dashboard">
  <div class="card">
    <div class="card-title">📊 Top Cache Usage (Data + Indexes)</div>
    <div class="chart-container" id="topCaches">
      <div class="loading-placeholder">
        <div class="spinner"></div>
        <div class="loading-text">Loading cache data...</div>
      </div>
    </div>
  </div>
  
  <div class="card">
    <div class="card-title">📈 Cache Distribution</div>
    <div class="chart-container" id="cacheDistr">
      <div class="loading-placeholder">
        <div class="spinner"></div>
        <div class="loading-text">Loading cache data...</div>
      </div>
    </div>
  </div>
  
  <div class="card">
    <div class="card-title">💾 Top Data in Cache</div>
    <div class="chart-container" id="topDataCache">
      <div class="loading-placeholder">
        <div class="spinner"></div>
        <div class="loading-text">Loading cache data...</div>
      </div>
    </div>
  </div>
  
  <div class="card">
    <div class="card-title">🔑 Top Indexes in Cache</div>
    <div class="chart-container" id="topIndexesCache">
      <div class="loading-placeholder">
        <div class="spinner"></div>
        <div class="loading-text">Loading cache data...</div>
      </div>
    </div>
  </div>
</div>

<div class="footer">
  Powered by <a href="https://github.com/simagix/keyhole" target="_blank">Keyhole</a> • 
  Refreshes every 5 seconds
</div>

<script>
  google.charts.load('current', {'packages':['corechart']});
  google.charts.setOnLoadCallback(drawCharts);

  const chartColors = ['#58a6ff', '#3fb950', '#a371f7', '#d29922', '#f85149', '#79c0ff', '#7ee787', '#d2a8ff', '#e3b341', '#ffa198'];
  const bgColor = '#21262d';
  const textColor = '#e6edf3';
  const gridColor = '#30363d';
  let dataReceived = false;

  function drawCharts() {
    redraw();
    setInterval(redraw, 5000);
  }

  function updateStatus(success) {
    const dot = document.getElementById('statusDot');
    const text = document.getElementById('lastUpdate');
    if (success) {
      dot.classList.remove('loading');
      const now = new Date();
      text.textContent = 'Last update: ' + now.toLocaleTimeString();
    } else {
      dot.classList.add('loading');
      text.textContent = 'Connecting...';
    }
  }

  function showNoData(divID, message) {
    document.getElementById(divID).innerHTML = 
      '<div class="no-data">' +
      '<div class="no-data-icon">📭</div>' +
      '<div>' + message + '</div>' +
      '</div>';
  }

  function hasValidData(data) {
    return data && data.length > 1 && data[1] && data[1][1] !== undefined && data[1][1] !== 0;
  }

  function redraw() {
    var xmlhttp = new XMLHttpRequest();
    var url = "wt/data";
    xmlhttp.onreadystatechange = function() {
      if (this.readyState == 4) {
        if (this.status == 200) {
          try {
            var doc = JSON.parse(this.responseText);
            dataReceived = true;
            
            if (hasValidData(doc.top_caches.data)) {
              drawPieChart('topCaches', doc.top_caches.title, doc.top_caches.data);
            } else {
              showNoData('topCaches', 'No cache data yet');
            }
            
            if (hasValidData(doc.cache_distr.data)) {
              drawPieChart('cacheDistr', doc.cache_distr.title, doc.cache_distr.data);
            } else {
              showNoData('cacheDistr', 'No cache data yet');
            }
            
            if (hasValidData(doc.top_data_cache.data)) {
              drawColumnChart('topDataCache', doc.top_data_cache.title, doc.top_data_cache.data);
            } else {
              showNoData('topDataCache', 'No data in cache yet');
            }
            
            if (hasValidData(doc.top_indexes_cache.data)) {
              drawColumnChart('topIndexesCache', doc.top_indexes_cache.title, doc.top_indexes_cache.data);
            } else {
              showNoData('topIndexesCache', 'No indexes in cache yet');
            }
            
            updateStatus(true);
          } catch (e) {
            console.error('Parse error:', e);
            updateStatus(false);
          }
        } else {
          updateStatus(false);
        }
      }
    };
    xmlhttp.open("GET", url, true);
    xmlhttp.send();
  }

  function drawPieChart(divID, title, data) {
    var chart_data = new google.visualization.arrayToDataTable(data);
    var options = {
      backgroundColor: bgColor,
      width: '100%',
      height: 320,
      pieHole: 0.4,
      colors: chartColors,
      legend: {
        position: 'right',
        textStyle: { color: textColor, fontSize: 11 }
      },
      pieSliceTextStyle: { color: textColor },
      chartArea: { width: '90%', height: '85%' },
      tooltip: { textStyle: { fontSize: 12 } }
    };
    var chart = new google.visualization.PieChart(document.getElementById(divID));
    chart.draw(chart_data, options);
  }

  function drawColumnChart(divID, title, data) {
    var chart_data = new google.visualization.arrayToDataTable(data);
    var options = {
      backgroundColor: bgColor,
      width: '100%',
      height: 320,
      colors: [chartColors[0]],
      legend: { position: 'none' },
      chartArea: { width: '85%', height: '75%' },
      hAxis: {
        textStyle: { color: textColor, fontSize: 10 },
        slantedText: true,
        slantedTextAngle: 45,
        gridlines: { color: gridColor }
      },
      vAxis: {
        textStyle: { color: textColor },
        gridlines: { color: gridColor },
        minorGridlines: { count: 0 }
      },
      bar: { groupWidth: '70%' },
      tooltip: { textStyle: { fontSize: 12 } }
    };
    var chart = new google.visualization.ColumnChart(document.getElementById(divID));
    chart.draw(chart_data, options);
  }
</script>
</body>
</html>
`
