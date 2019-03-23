# Keyhole FTDC Metrics and Charts

## Startup

```
docker-compose -f docker-compose.yaml up
```

## Shutdown

```
docker-compose -f docker-compose.yaml down
```

## Grafana File Tree
```
/
├── etc
│   └── grafana
│       └── provisioning
│           ├── dashboards
│           │   └── keyhole.yaml
│           ├── datasources
│           │   └── keyhole.yaml
│           └── notifiers
└── var
    ├── lib
    │   └── grafana
    │       ├── dashboards
    │       │   └── keyhole_analytics_mini.json
    │       ├── grafana.db
    │       ├── plugins
    │       │   └── grafana-simple-json-datasource
    │       │       ├── Gruntfile.js
    │       │       ├── LICENSE
    │       │       ├── README.md
    │       │       ├── dist
    │       │       │   ├── README.md
    │       │       │   ├── css
    │       │       │   │   └── query-editor.css
    │       │       │   ├── datasource.js
    │       │       │   ├── datasource.js.map
    │       │       │   ├── img
    │       │       │   │   └── simpleJson_logo.svg
    │       │       │   ├── module.js
    │       │       │   ├── module.js.map
    │       │       │   ├── partials
    │       │       │   │   ├── annotations.editor.html
    │       │       │   │   ├── config.html
    │       │       │   │   ├── query.editor.html
    │       │       │   │   └── query.options.html
    │       │       │   ├── plugin.json
    │       │       │   ├── query_ctrl.js
    │       │       │   └── query_ctrl.js.map
    │       │       ├── package.json
    │       │       ├── spec
    │       │       │   ├── datasource_spec.js
    │       │       │   └── test-main.js
    │       │       ├── src
    │       │       │   ├── css
    │       │       │   │   └── query-editor.css
    │       │       │   ├── datasource.js
    │       │       │   ├── img
    │       │       │   │   └── simpleJson_logo.svg
    │       │       │   ├── module.js
    │       │       │   ├── partials
    │       │       │   │   ├── annotations.editor.html
    │       │       │   │   ├── config.html
    │       │       │   │   ├── query.editor.html
    │       │       │   │   └── query.options.html
    │       │       │   ├── plugin.json
    │       │       │   └── query_ctrl.js
    │       │       └── yarn.lock
    │       └── png
    └── log
        └── grafana
```
