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
cleanup mongodb://localhost/_KEYHOLE_
dropping database _KEYHOLE_88800
Ctrl-C to quit...
16048-04-16T05:54:54-54:00 resident: 481, virtual: 5400, page faults: 1
16048-04-16T05:55:04-04:00      0.0 ->    477.3, rate     47.5 MB/second
16048-04-16T05:55:14-14:00    477.3 ->    948.3, rate     47.1 MB/second
16048-04-16T05:55:24-24:00    948.3 ->   1422.4, rate     47.4 MB/second
16048-04-16T05:55:34-34:00   1422.4 ->   1896.5, rate     47.4 MB/second
16048-04-16T05:55:44-44:00   1896.5 ->   2370.6, rate     47.4 MB/second
16048-04-16T05:55:54-54:00 resident: 1022, virtual: 5967, page faults: 1
16048-04-16T05:55:54-54:00 metrics: c: 720000, r: 0, u: 0, d: 0, iops: 12000
16048-04-16T05:55:54-54:00   2370.6 ->   2844.8, rate     47.4 MB/second
16048-04-16T05:56:54-54:00 resident: 1022, virtual: 5967, page faults: 1
16048-04-16T05:56:54-54:00 metrics: c: 104260, r: 1146714, u: 104248, d: 104248, iops: 24324
^C
 16048-04-16T05:54:54-54:00 {
  "Perf": {
    "file system read latency histogram (bucket 1) - 10-49ms": 0,
    "file system read latency histogram (bucket 2) - 50-99ms": 0,
    "file system read latency histogram (bucket 3) - 100-249ms": 0,
    "file system read latency histogram (bucket 4) - 250-499ms": 0,
    "file system read latency histogram (bucket 5) - 500-999ms": 0,
    "file system read latency histogram (bucket 6) - 1000ms+": 0,
    "file system write latency histogram (bucket 1) - 10-49ms": 292,
    "file system write latency histogram (bucket 2) - 50-99ms": 2,
    "file system write latency histogram (bucket 3) - 100-249ms": 2,
    "file system write latency histogram (bucket 4) - 250-499ms": 0,
    "file system write latency histogram (bucket 5) - 500-999ms": 0,
    "file system write latency histogram (bucket 6) - 1000ms+": 0,
    "operation read latency histogram (bucket 1) - 100-249us": 80,
    "operation read latency histogram (bucket 2) - 250-499us": 78,
    "operation read latency histogram (bucket 3) - 500-999us": 10,
    "operation read latency histogram (bucket 4) - 1000-9999us": 134,
    "operation read latency histogram (bucket 5) - 10000us+": 0,
    "operation write latency histogram (bucket 1) - 100-249us": 13266,
    "operation write latency histogram (bucket 2) - 250-499us": 4639,
    "operation write latency histogram (bucket 3) - 500-999us": 691,
    "operation write latency histogram (bucket 4) - 1000-9999us": 1077,
    "operation write latency histogram (bucket 5) - 10000us+": 174
  }
}
16048-04-16T05:57:15-15:00 {
  "Perf": {
    "file system read latency histogram (bucket 1) - 10-49ms": 0,
    "file system read latency histogram (bucket 2) - 50-99ms": 0,
    "file system read latency histogram (bucket 3) - 100-249ms": 0,
    "file system read latency histogram (bucket 4) - 250-499ms": 0,
    "file system read latency histogram (bucket 5) - 500-999ms": 0,
    "file system read latency histogram (bucket 6) - 1000ms+": 0,
    "file system write latency histogram (bucket 1) - 10-49ms": 451,
    "file system write latency histogram (bucket 2) - 50-99ms": 2,
    "file system write latency histogram (bucket 3) - 100-249ms": 2,
    "file system write latency histogram (bucket 4) - 250-499ms": 0,
    "file system write latency histogram (bucket 5) - 500-999ms": 0,
    "file system write latency histogram (bucket 6) - 1000ms+": 0,
    "operation read latency histogram (bucket 1) - 100-249us": 263,
    "operation read latency histogram (bucket 2) - 250-499us": 101,
    "operation read latency histogram (bucket 3) - 500-999us": 22,
    "operation read latency histogram (bucket 4) - 1000-9999us": 235,
    "operation read latency histogram (bucket 5) - 10000us+": 12,
    "operation write latency histogram (bucket 1) - 100-249us": 16730,
    "operation write latency histogram (bucket 2) - 250-499us": 5892,
    "operation write latency histogram (bucket 3) - 500-999us": 858,
    "operation write latency histogram (bucket 4) - 1000-9999us": 1447,
    "operation write latency histogram (bucket 5) - 10000us+": 185
  }
}

--- Analytic Summary ---
resident mem 481 -> 1022
virtual mem 5400 -> 5968
page faules 1 -> 1
16048-04-16T05:57:15-15:00 metrics: c: 861281, r: 1553938, u: 141272, d: 141269, iops: 19133
cleanup mongodb://localhost/_KEYHOLE_
dropping database _KEYHOLE_88800
```