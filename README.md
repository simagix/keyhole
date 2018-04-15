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
2018-04-15T19:39:55-04:00 resident: 890, virtual: 5912, page faults: 12
2018-04-15T19:39:55-04:00 metrics: c: 1334743, r: 9603830, u: 614726, d: 614723, iops: 0
2018-04-15T19:40:05-04:00      0.0 ->    465.4, rate     45.7 MB/second
2018-04-15T19:40:15-04:00    465.4 ->    872.9, rate     40.5 MB/second
2018-04-15T19:40:25-04:00    872.9 ->   1198.1, rate     32.4 MB/second
2018-04-15T19:40:35-04:00   1198.1 ->   1492.9, rate     29.4 MB/second
2018-04-15T19:40:45-04:00   1492.9 ->   1846.9, rate     35.3 MB/second
2018-04-15T19:40:55-04:00 resident: 866, virtual: 5931, page faults: 12
2018-04-15T19:40:55-04:00 metrics: c: 1901285, r: 10170253, u: 614726, d: 614723, iops: 18882
2018-04-15T19:40:55-04:00   1846.9 ->   2244.1, rate     39.6 MB/second
2018-04-15T19:41:05-04:00   2244.1 ->   2527.1, rate     28.2 MB/second
2018-04-15T19:41:15-04:00   2527.1 ->   2844.5, rate     31.5 MB/second
2018-04-15T19:41:25-04:00   2844.5 ->   2844.8, rate      0.0 MB/second
2018-04-15T19:41:55-04:00 resident: 895, virtual: 5981, page faults: 12
2018-04-15T19:41:55-04:00 metrics: c: 2121042, r: 11251803, u: 681024, d: 681017, iops: 23898
^C
Key: 2018-04-15T19:39:55-04:00 Value: {
  "Mem": {
    "resident": 890,
    "virtual": 5912
  },
  "extra_info": {
    "page_faults": 12
  },
  "metrics": {
    "document": {
      "deleted": 614723,
      "inserted": 1334743,
      "returned": 9603830,
      "updated": 614726
    }
  },
  "WiredTiger": {
    "Perf": {
      "file system read latency histogram (bucket 1) - 10-49ms": 0,
      "file system read latency histogram (bucket 2) - 50-99ms": 0,
      "file system read latency histogram (bucket 3) - 100-249ms": 0,
      "file system read latency histogram (bucket 4) - 250-499ms": 0,
      "file system read latency histogram (bucket 5) - 500-999ms": 0,
      "file system read latency histogram (bucket 6) - 1000ms+": 0,
      "file system write latency histogram (bucket 1) - 10-49ms": 24445,
      "file system write latency histogram (bucket 2) - 50-99ms": 38,
      "file system write latency histogram (bucket 3) - 100-249ms": 22,
      "file system write latency histogram (bucket 4) - 250-499ms": 32,
      "file system write latency histogram (bucket 5) - 500-999ms": 10,
      "file system write latency histogram (bucket 6) - 1000ms+": 3,
      "operation read latency histogram (bucket 1) - 100-249us": 6854,
      "operation read latency histogram (bucket 2) - 250-499us": 2809,
      "operation read latency histogram (bucket 3) - 500-999us": 995,
      "operation read latency histogram (bucket 4) - 1000-9999us": 1091,
      "operation read latency histogram (bucket 5) - 10000us+": 376,
      "operation write latency histogram (bucket 1) - 100-249us": 7064,
      "operation write latency histogram (bucket 2) - 250-499us": 3771,
      "operation write latency histogram (bucket 3) - 500-999us": 2301,
      "operation write latency histogram (bucket 4) - 1000-9999us": 5900,
      "operation write latency histogram (bucket 5) - 10000us+": 3074
    }
  }
}
Key: 2018-04-15T19:42:05-04:00 Value: {
  "Mem": {
    "resident": 901,
    "virtual": 5981
  },
  "extra_info": {
    "page_faults": 12
  },
  "metrics": {
    "document": {
      "deleted": 696858,
      "inserted": 2136884,
      "returned": 11473638,
      "updated": 696866
    }
  },
  "WiredTiger": {
    "Perf": {
      "file system read latency histogram (bucket 1) - 10-49ms": 0,
      "file system read latency histogram (bucket 2) - 50-99ms": 0,
      "file system read latency histogram (bucket 3) - 100-249ms": 0,
      "file system read latency histogram (bucket 4) - 250-499ms": 0,
      "file system read latency histogram (bucket 5) - 500-999ms": 0,
      "file system read latency histogram (bucket 6) - 1000ms+": 0,
      "file system write latency histogram (bucket 1) - 10-49ms": 44557,
      "file system write latency histogram (bucket 2) - 50-99ms": 47,
      "file system write latency histogram (bucket 3) - 100-249ms": 36,
      "file system write latency histogram (bucket 4) - 250-499ms": 44,
      "file system write latency histogram (bucket 5) - 500-999ms": 10,
      "file system write latency histogram (bucket 6) - 1000ms+": 7,
      "operation read latency histogram (bucket 1) - 100-249us": 8618,
      "operation read latency histogram (bucket 2) - 250-499us": 3526,
      "operation read latency histogram (bucket 3) - 500-999us": 1347,
      "operation read latency histogram (bucket 4) - 1000-9999us": 1551,
      "operation read latency histogram (bucket 5) - 10000us+": 696,
      "operation write latency histogram (bucket 1) - 100-249us": 13434,
      "operation write latency histogram (bucket 2) - 250-499us": 7571,
      "operation write latency histogram (bucket 3) - 500-999us": 4926,
      "operation write latency histogram (bucket 4) - 1000-9999us": 11371,
      "operation write latency histogram (bucket 5) - 10000us+": 5950
    }
  }
}
cleanup mongodb://localhost/_KEYHOLE_?replicaSet=replset
dropping database _KEYHOLE_88800
```