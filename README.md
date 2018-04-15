# keyhole

Peek into `mongod` for

- Load tests
- Seed data
- Collect server status


## Usage
```
$ build/keyhole-osx-x64 -h
Usage of build/keyhole-osx-x64:
  -info
    	get cluster info
  -seed
    	seed a database for demo
  -uri string
    	MongoDB URI (default "mongodb://localhost")
  -v	verbose
```

## Example
```
build/keyhole-osx-x64 -uri=mongodb://localhost/_KEYHOLE_?replicaSet=replset

MongoDB URI: mongodb://localhost/_KEYHOLE_?replicaSet=replset
info: false
seed: false
cleanup mongodb://localhost/_KEYHOLE_?replicaSet=replset
dropping database _KEYHOLE_88800
Ctrl-C to quit...
2018-04-15T08:51:50-04:00      0.0 ->    393.0, rate     39.3 MB/second
2018-04-15T08:52:00-04:00    393.0 ->    700.6, rate     30.7 MB/second
2018-04-15T08:52:10-04:00    700.6 ->   1014.8, rate     31.4 MB/second
2018-04-15T08:52:20-04:00   1014.8 ->   1514.9, rate     49.8 MB/second
2018-04-15T08:52:30-04:00   1514.9 ->   2006.1, rate     48.9 MB/second
^CKey: 2018-04-15T08:51:40-04:00 Value: {
  "Mem": {
    "bits": 64,
    "mapped": 0,
    "mappedWithJournal": 0,
    "resident": 1054,
    "supported": true,
    "virtual": 6064
  },
  "extra_info": {
    "note": "fields vary by platform",
    "page_faults": 566
  },
  "WiredTiger": {
    "Perf": {
      "file system read latency histogram (bucket 1) - 10-49ms": 43,
      "file system read latency histogram (bucket 2) - 50-99ms": 0,
      "file system read latency histogram (bucket 3) - 100-249ms": 0,
      "file system read latency histogram (bucket 4) - 250-499ms": 0,
      "file system read latency histogram (bucket 5) - 500-999ms": 0,
      "file system read latency histogram (bucket 6) - 1000ms+": 0,
      "file system write latency histogram (bucket 1) - 10-49ms": 238079,
      "file system write latency histogram (bucket 2) - 50-99ms": 192,
      "file system write latency histogram (bucket 3) - 100-249ms": 242,
      "file system write latency histogram (bucket 4) - 250-499ms": 63,
      "file system write latency histogram (bucket 5) - 500-999ms": 107,
      "file system write latency histogram (bucket 6) - 1000ms+": 0,
      "operation read latency histogram (bucket 1) - 100-249us": 2137,
      "operation read latency histogram (bucket 2) - 250-499us": 921,
      "operation read latency histogram (bucket 3) - 500-999us": 466,
      "operation read latency histogram (bucket 4) - 1000-9999us": 171,
      "operation read latency histogram (bucket 5) - 10000us+": 44,
      "operation write latency histogram (bucket 1) - 100-249us": 327175,
      "operation write latency histogram (bucket 2) - 250-499us": 151417,
      "operation write latency histogram (bucket 3) - 500-999us": 101927,
      "operation write latency histogram (bucket 4) - 1000-9999us": 78675,
      "operation write latency histogram (bucket 5) - 10000us+": 2778
    }
  }
}
Key: 2018-04-15T08:52:36-04:00 Value: {
  "Mem": {
    "bits": 64,
    "mapped": 0,
    "mappedWithJournal": 0,
    "resident": 840,
    "supported": true,
    "virtual": 5947
  },
  "extra_info": {
    "note": "fields vary by platform",
    "page_faults": 997
  },
  "WiredTiger": {
    "Perf": {
      "file system read latency histogram (bucket 1) - 10-49ms": 223,
      "file system read latency histogram (bucket 2) - 50-99ms": 0,
      "file system read latency histogram (bucket 3) - 100-249ms": 0,
      "file system read latency histogram (bucket 4) - 250-499ms": 0,
      "file system read latency histogram (bucket 5) - 500-999ms": 0,
      "file system read latency histogram (bucket 6) - 1000ms+": 0,
      "file system write latency histogram (bucket 1) - 10-49ms": 166476,
      "file system write latency histogram (bucket 2) - 50-99ms": 439,
      "file system write latency histogram (bucket 3) - 100-249ms": 207,
      "file system write latency histogram (bucket 4) - 250-499ms": 88,
      "file system write latency histogram (bucket 5) - 500-999ms": 40,
      "file system write latency histogram (bucket 6) - 1000ms+": 9,
      "operation read latency histogram (bucket 1) - 100-249us": 12739,
      "operation read latency histogram (bucket 2) - 250-499us": 4838,
      "operation read latency histogram (bucket 3) - 500-999us": 2308,
      "operation read latency histogram (bucket 4) - 1000-9999us": 3947,
      "operation read latency histogram (bucket 5) - 10000us+": 2363,
      "operation write latency histogram (bucket 1) - 100-249us": 62777,
      "operation write latency histogram (bucket 2) - 250-499us": 34922,
      "operation write latency histogram (bucket 3) - 500-999us": 16611,
      "operation write latency histogram (bucket 4) - 1000-9999us": 25898,
      "operation write latency histogram (bucket 5) - 10000us+": 12706
    }
  }
}
cleanup mongodb://localhost/_KEYHOLE_?replicaSet=replset
dropping database _KEYHOLE_88800
```