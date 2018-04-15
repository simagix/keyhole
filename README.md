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
2018-04-15T14:01:30-04:00 resident: 35, virtual: 5144, page faults: 393
2018-04-15T14:01:40-04:00      0.1 ->    474.6, rate     47.2 MB/second
2018-04-15T14:01:50-04:00    474.6 ->    922.9, rate     44.8 MB/second
2018-04-15T14:02:00-04:00    922.9 ->   1239.9, rate     31.6 MB/second
2018-04-15T14:02:10-04:00   1239.9 ->   1576.2, rate     33.6 MB/second
2018-04-15T14:02:20-04:00   1576.2 ->   1893.2, rate     31.6 MB/second
2018-04-15T14:02:30-04:00 resident: 790, virtual: 5848, page faults: 393
2018-04-15T14:02:30-04:00   1893.2 ->   2225.2, rate     33.1 MB/second
2018-04-15T14:02:40-04:00   2225.2 ->   2634.3, rate     40.8 MB/second
2018-04-15T14:02:50-04:00   2634.3 ->   2844.8, rate     20.9 MB/second
2018-04-15T14:03:30-04:00 resident: 816, virtual: 5891, page faults: 393
2018-04-15T14:04:30-04:00 resident: 840, virtual: 5914, page faults: 393
2018-04-15T14:05:30-04:00 resident: 856, virtual: 5930, page faults: 393
2018-04-15T14:06:30-04:00 resident: 863, virtual: 5934, page faults: 393
^C
Key: 2018-04-15T14:01:30-04:00 Value: {
  "Mem": {
    "resident": 35,
    "virtual": 5144
  },
  "extra_info": {
    "page_faults": 393
  },
  "WiredTiger": {
    "Perf": {
      "file system read latency histogram (bucket 1) - 10-49ms": 0,
      "file system read latency histogram (bucket 2) - 50-99ms": 0,
      "file system read latency histogram (bucket 3) - 100-249ms": 0,
      "file system read latency histogram (bucket 4) - 250-499ms": 0,
      "file system read latency histogram (bucket 5) - 500-999ms": 0,
      "file system read latency histogram (bucket 6) - 1000ms+": 0,
      "file system write latency histogram (bucket 1) - 10-49ms": 0,
      "file system write latency histogram (bucket 2) - 50-99ms": 0,
      "file system write latency histogram (bucket 3) - 100-249ms": 0,
      "file system write latency histogram (bucket 4) - 250-499ms": 0,
      "file system write latency histogram (bucket 5) - 500-999ms": 0,
      "file system write latency histogram (bucket 6) - 1000ms+": 0,
      "operation read latency histogram (bucket 1) - 100-249us": 1,
      "operation read latency histogram (bucket 2) - 250-499us": 1,
      "operation read latency histogram (bucket 3) - 500-999us": 0,
      "operation read latency histogram (bucket 4) - 1000-9999us": 0,
      "operation read latency histogram (bucket 5) - 10000us+": 0,
      "operation write latency histogram (bucket 1) - 100-249us": 1,
      "operation write latency histogram (bucket 2) - 250-499us": 0,
      "operation write latency histogram (bucket 3) - 500-999us": 0,
      "operation write latency histogram (bucket 4) - 1000-9999us": 0,
      "operation write latency histogram (bucket 5) - 10000us+": 0
    }
  }
}
Key: 2018-04-15T14:07:16-04:00 Value: {
  "Mem": {
    "resident": 865,
    "virtual": 5932
  },
  "extra_info": {
    "page_faults": 393
  },
  "WiredTiger": {
    "Perf": {
      "file system read latency histogram (bucket 1) - 10-49ms": 4,
      "file system read latency histogram (bucket 2) - 50-99ms": 0,
      "file system read latency histogram (bucket 3) - 100-249ms": 0,
      "file system read latency histogram (bucket 4) - 250-499ms": 0,
      "file system read latency histogram (bucket 5) - 500-999ms": 0,
      "file system read latency histogram (bucket 6) - 1000ms+": 0,
      "file system write latency histogram (bucket 1) - 10-49ms": 21352,
      "file system write latency histogram (bucket 2) - 50-99ms": 84,
      "file system write latency histogram (bucket 3) - 100-249ms": 46,
      "file system write latency histogram (bucket 4) - 250-499ms": 10,
      "file system write latency histogram (bucket 5) - 500-999ms": 6,
      "file system write latency histogram (bucket 6) - 1000ms+": 6,
      "operation read latency histogram (bucket 1) - 100-249us": 5050,
      "operation read latency histogram (bucket 2) - 250-499us": 2033,
      "operation read latency histogram (bucket 3) - 500-999us": 829,
      "operation read latency histogram (bucket 4) - 1000-9999us": 928,
      "operation read latency histogram (bucket 5) - 10000us+": 385,
      "operation write latency histogram (bucket 1) - 100-249us": 7054,
      "operation write latency histogram (bucket 2) - 250-499us": 4115,
      "operation write latency histogram (bucket 3) - 500-999us": 2660,
      "operation write latency histogram (bucket 4) - 1000-9999us": 6112,
      "operation write latency histogram (bucket 5) - 10000us+": 3177
    }
  }
}
cleanup mongodb://localhost/_KEYHOLE_?replicaSet=replset
dropping database _KEYHOLE_88800
```