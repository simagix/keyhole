# keyhole

Peek into `mongod` for

- Load tests
- Seed data
- Collect server status


## Usage
```
$ build/keyhole-osx-x64 -h
  -conn int
    	nuumber of connections (default 20)
  -info
    	get cluster info
  -seed
    	seed a database for demo
  -tps int
    	number of trasaction per second per connection (default 600)
  -uri string
    	MongoDB URI (default "mongodb://localhost")
  -v	verbose
```

## Example
```
build/keyhole-osx-x64 -uri=mongodb://localhost/_KEYHOLE_?replicaSet=replset

MongoDB URI: mongodb://localhost/_KEYHOLE_?replicaSet=replset
Total TPS: 600 (tps) * 20 (conns) = 12000
cleanup mongodb://localhost/_KEYHOLE_?replicaSet=replset
dropping database _KEYHOLE_88800
Ctrl-C to quit...
2018-04-15T09:52:16-04:00      0.1 ->    476.5, rate     47.4 MB/second
2018-04-15T09:52:26-04:00    476.5 ->    948.7, rate     47.2 MB/second
2018-04-15T09:52:36-04:00    948.7 ->   1422.4, rate     47.3 MB/second
2018-04-15T09:52:46-04:00   1422.4 ->   1896.5, rate     47.4 MB/second
2018-04-15T09:52:56-04:00   1896.5 ->   2247.5, rate     35.1 MB/second
2018-04-15T09:53:06-04:00   2247.5 ->   2615.1, rate     36.6 MB/second
^C
Key: 2018-04-15T09:52:06-04:00 Value: {
  "Mem": {
    "bits": 64,
    "mapped": 0,
    "mappedWithJournal": 0,
    "resident": 1058,
    "supported": true,
    "virtual": 6122
  },
  "extra_info": {
    "note": "fields vary by platform",
    "page_faults": 1014
  },
  "WiredTiger": {
    "Perf": {
      "file system read latency histogram (bucket 1) - 10-49ms": 224,
      "file system read latency histogram (bucket 2) - 50-99ms": 0,
      "file system read latency histogram (bucket 3) - 100-249ms": 0,
      "file system read latency histogram (bucket 4) - 250-499ms": 0,
      "file system read latency histogram (bucket 5) - 500-999ms": 0,
      "file system read latency histogram (bucket 6) - 1000ms+": 0,
      "file system write latency histogram (bucket 1) - 10-49ms": 236872,
      "file system write latency histogram (bucket 2) - 50-99ms": 537,
      "file system write latency histogram (bucket 3) - 100-249ms": 260,
      "file system write latency histogram (bucket 4) - 250-499ms": 131,
      "file system write latency histogram (bucket 5) - 500-999ms": 50,
      "file system write latency histogram (bucket 6) - 1000ms+": 16,
      "operation read latency histogram (bucket 1) - 100-249us": 19348,
      "operation read latency histogram (bucket 2) - 250-499us": 7263,
      "operation read latency histogram (bucket 3) - 500-999us": 3389,
      "operation read latency histogram (bucket 4) - 1000-9999us": 5140,
      "operation read latency histogram (bucket 5) - 10000us+": 2947,
      "operation write latency histogram (bucket 1) - 100-249us": 96746,
      "operation write latency histogram (bucket 2) - 250-499us": 57627,
      "operation write latency histogram (bucket 3) - 500-999us": 28072,
      "operation write latency histogram (bucket 4) - 1000-9999us": 40199,
      "operation write latency histogram (bucket 5) - 10000us+": 19474
    }
  }
}
Key: 2018-04-15T09:53:06-04:00 Value: {
  "Mem": {
    "bits": 64,
    "mapped": 0,
    "mappedWithJournal": 0,
    "resident": 1066,
    "supported": true,
    "virtual": 6158
  },
  "extra_info": {
    "note": "fields vary by platform",
    "page_faults": 1014
  },
  "WiredTiger": {
    "Perf": {
      "file system read latency histogram (bucket 1) - 10-49ms": 224,
      "file system read latency histogram (bucket 2) - 50-99ms": 0,
      "file system read latency histogram (bucket 3) - 100-249ms": 0,
      "file system read latency histogram (bucket 4) - 250-499ms": 0,
      "file system read latency histogram (bucket 5) - 500-999ms": 0,
      "file system read latency histogram (bucket 6) - 1000ms+": 0,
      "file system write latency histogram (bucket 1) - 10-49ms": 246444,
      "file system write latency histogram (bucket 2) - 50-99ms": 542,
      "file system write latency histogram (bucket 3) - 100-249ms": 262,
      "file system write latency histogram (bucket 4) - 250-499ms": 133,
      "file system write latency histogram (bucket 5) - 500-999ms": 51,
      "file system write latency histogram (bucket 6) - 1000ms+": 16,
      "operation read latency histogram (bucket 1) - 100-249us": 19931,
      "operation read latency histogram (bucket 2) - 250-499us": 7459,
      "operation read latency histogram (bucket 3) - 500-999us": 3484,
      "operation read latency histogram (bucket 4) - 1000-9999us": 5285,
      "operation read latency histogram (bucket 5) - 10000us+": 3062,
      "operation write latency histogram (bucket 1) - 100-249us": 100388,
      "operation write latency histogram (bucket 2) - 250-499us": 60288,
      "operation write latency histogram (bucket 3) - 500-999us": 29492,
      "operation write latency histogram (bucket 4) - 1000-9999us": 41918,
      "operation write latency histogram (bucket 5) - 10000us+": 20714
    }
  }
}
cleanup mongodb://localhost/_KEYHOLE_?replicaSet=replset
dropping database _KEYHOLE_88800
```