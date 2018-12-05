# Keyhole - Load Test
Load test a cluster/replica.  A default cycle last six minutes.

- Populate data in first minute
- Perform CRUD operations during the second minutes
- Burst test until before the last minute
- Perform cleanup ops in the last minute

## Write Throughputs Test
Measure MongoDB write throughputs.

```
keyhole --duration 1 mongodb://localhost/?replicaSet=replset
```

By default, it writes 2K size documents at 60 transactions per second from 10 different threads, a total of 600 TPS.  See sample outputs below.

```
Duration in minute(s): 5
...
Total TPS: 300 (tps) * 10 (conns) = 3000, duration: 5 (mins), bulk size: 512
...

2018-06-10T14:40:25-04:00 [replset] Memory - resident: 1067, virtual: 6100
2018-06-10T14:40:36-04:00 [replset] Storage: 460.6 -> 809.3, rate: 34.8 MB/sec
2018-06-10T14:40:46-04:00 [replset] Storage: 809.3 -> 1091.8, rate: 28.1 MB/sec
2018-06-10T14:40:56-04:00 [replset] Storage: 1091.8 -> 1375.6, rate: 28.2 MB/sec
2018-06-10T14:41:06-04:00 [replset] Storage: 1375.6 -> 1662.1, rate: 28.2 MB/sec
2018-06-10T14:41:16-04:00 [replset] Storage: 1662.1 -> 1869.0, rate: 20.6 MB/sec
```

## Load Test
Keyhole uses a template to read in a sample document and then generates randomized documents based on the template (see [wiki](https://github.com/simagix/keyhole/wiki/Seed-Data-using-a-Template) for details.  To execute a load test with a template, do

```
keyhole --file example/template.json "mongodb://localhost/?replicaSet=replset"
```

To execute a demo load test, run it without the `--file` flag.

```
keyhole "mongodb://localhost/?replicaSet=replset"
```

## Document Example
### Usage
```
keyhole -schema
```

### Schema
```
{
  "favoriteBook": "Ulysses",
  "favoriteBooks": [
    "Ulysses",
    "The Lord of the Rings",
    "In Search of Lost Time"
  ],
  "favoriteCities": [
    "Atlanta",
    "New York",
    "Bangkok"
  ],
  "favoriteCity": "Atlanta",
  "favoriteMovie": "Lawrence of Arabia",
  "favoriteMovies": [
    "Lawrence of Arabia",
    "One Flew Over the Cuckoo's Nest",
    "Casablanca"
  ],
  "favoriteMusic": "Classical",
  "favoriteMusics": [
    "Classical",
    "Soul",
    "Hip Pop"
  ],
  "favoriteSport": "Skateboard",
  "favoriteSports": [
    "Skateboard",
    "Soccer",
    "Baseball"
  ],
  "favorites": {
    "book": "Ulysses",
    "books": [
      "Ulysses",
      "The Lord of the Rings",
      "In Search of Lost Time"
    ],
    "cities": [
      "Atlanta",
      "New York",
      "Bangkok"
    ],
    "city": "Atlanta",
    "movie": "Lawrence of Arabia",
    "movies": [
      "Lawrence of Arabia",
      "One Flew Over the Cuckoo's Nest",
      "Casablanca"
    ],
    "music": "Classical",
    "musics": [
      "Classical",
      "Soul",
      "Hip Pop"
    ],
    "sport": "Skateboard",
    "sports": [
      "Skateboard",
      "Soccer",
      "Baseball"
    ]
  },
  "favoritesList": [
    {
      "book": "Ulysses",
      "city": "Atlanta",
      "movie": "Lawrence of Arabia",
      "music": "Classical",
      "sport": "Skateboard"
    },
    {
      "book": "The Lord of the Rings",
      "city": "New York",
      "movie": "One Flew Over the Cuckoo's Nest",
      "music": "Soul",
      "sport": "Soccer"
    },
    {
      "book": "In Search of Lost Time",
      "city": "Bangkok",
      "movie": "Casablanca",
      "music": "Hip Pop",
      "sport": "Baseball"
    }
  ],
  "filler1": "",
  "filler2": "",
  "number": 211,
  "numbers": [
    445,
    237,
    106,
    495,
    466
  ],
  "ts": "2018-05-27T12:16:44.506993-04:00"
}
```

## Example
```
$ keyhole --drop --cleanup mongodb://localhost/?replicaSet=replset

MongoDB URI: mongodb://localhost/?replicaSet=replset
Duration in minute(s): 5
2018/06/10 14:40:24 cleanup mongodb://localhost/?replicaSet=replset
2018/06/10 14:40:24 dropping collection _KEYHOLE_88800 examples
2018/06/10 14:40:24 dropping database _KEYHOLE_88800
Total TPS: 300 (tps) * 10 (conns) = 3000, duration: 5 (mins), bulk size: 512
CollectServerStatus: connect to replset
CollectDBStats: connect to replset, _KEYHOLE_88800

2018-06-10T14:40:25-04:00 [replset] Memory - resident: 1067, virtual: 6100
2018-06-10T14:40:36-04:00 [replset] Storage: 460.6 -> 809.3, rate: 34.8 MB/sec
2018-06-10T14:40:46-04:00 [replset] Storage: 809.3 -> 1091.8, rate: 28.1 MB/sec
2018-06-10T14:40:56-04:00 [replset] Storage: 1091.8 -> 1375.6, rate: 28.2 MB/sec
2018-06-10T14:41:06-04:00 [replset] Storage: 1375.6 -> 1662.1, rate: 28.2 MB/sec
2018-06-10T14:41:16-04:00 [replset] Storage: 1662.1 -> 1869.0, rate: 20.6 MB/sec

2018-06-10T14:41:27-04:00 [replset] Memory - resident: 1069, virtual: 6095, page faults: 0, iops: 38736.1
2018-06-10T14:41:27-04:00 [replset] CRUD+  - insert: 1154949, find: 1735, update: 576, delete: 0, getmore: 2045, command: 2017
2018-06-10T14:41:27-04:00 [replset] Latency- read: 6.6, write: 144.5, command: 0.3 (ms)
2018-06-10T14:41:26-04:00 [replset] Storage: 1869.0 -> 1884.9, rate: 1.6 MB/sec
2018-06-10T14:41:36-04:00 [replset] Storage: 1884.9 -> 1901.1, rate: 1.6 MB/sec
2018-06-10T14:41:46-04:00 [replset] Storage: 1901.1 -> 1917.9, rate: 1.7 MB/sec
CollectDBStats exiting...

2018-06-10T14:42:27-04:00 [replset] Memory - resident: 1069, virtual: 6092, page faults: 0, iops: 26060.5
2018-06-10T14:42:27-04:00 [replset] CRUD+  - insert: 60139, find: 180422, update: 60142, delete: 0, getmore: 5642, command: 6352
2018-06-10T14:42:27-04:00 [replset] Latency- read: 0.2, write: 0.2, command: 0.1 (ms)

2018-06-10T14:43:27-04:00 [replset] Memory - resident: 1069, virtual: 6091, page faults: 0, iops: 25396.8
2018-06-10T14:43:27-04:00 [replset] CRUD+  - insert: 58611, find: 175824, update: 58607, delete: 0, getmore: 6380, command: 7256
2018-06-10T14:43:27-04:00 [replset] Latency- read: 0.2, write: 0.2, command: 0.1 (ms)

2018-06-10T14:44:27-04:00 [replset] Memory - resident: 1100, virtual: 6163, page faults: 0, iops: 25122.9
2018-06-10T14:44:27-04:00 [replset] CRUD+  - insert: 57969, find: 173924, update: 57975, delete: 10, getmore: 6365, command: 7488
2018-06-10T14:44:27-04:00 [replset] Latency- read: 0.2, write: 0.2, command: 0.1 (ms)

stats written to /var/folders/mv/q3097r9j5kxb59sg1btgf2s80000gp/T//keyhole_stats.2018-06-10T144024-replset
--- Host: Kens-MBP, version: 3.6.4 ---

--- Analytic Summary ---
+-------------------------+-------+-------+------+--------+--------+--------+--------+--------+--------+--------+
| Date/Time               | res   | virt  | fault| Command| Delete | Getmore| Insert | Query  | Update | iops   |
|-------------------------|-------+-------|------|--------|--------|--------|--------|--------|--------|--------|
|2018-06-10T14:41:27-04:00|   1069|   6095|     0|    2017|       0|    2045| 1154949|    1735|     576|   19038|
|2018-06-10T14:42:27-04:00|   1069|   6092|     0|    6352|       0|    5642|   60139|  180422|   60142|    5211|
|2018-06-10T14:43:27-04:00|   1069|   6091|     0|    7256|       0|    6380|   58611|  175824|   58607|    5111|
|2018-06-10T14:44:27-04:00|   1100|   6163|     0|    7488|      10|    6365|   57969|  173924|   57975|    5062|
|2018-06-10T14:45:24-04:00|   1156|   6213|     0|    1238|     480|    3784|       0|       1|       0|      96|
+-------------------------+-------+-------+------+--------+--------+--------+--------+--------+--------+--------+

--- Global Locks Summary ---
+-------------------------+--------------+--------------------------------------------+--------------------------------------------+
|                         | Total Time   | Active Clients                             | Current Queue                              |
| Date/Time               | (ms)         | total        | readers      | writers      | total        | readers      | writers      |
|-------------------------|--------------|--------------|--------------|--------------|--------------|--------------|--------------|
|2018-06-10T14:41:27-04:00|         61953|             0|             0|             0|             0|             0|             0|
|2018-06-10T14:42:27-04:00|         60375|             1|             1|             0|             1|             1|             0|
|2018-06-10T14:43:27-04:00|         60028|             0|             0|             0|             0|             0|             0|
|2018-06-10T14:44:27-04:00|         60013|             0|             0|             0|             0|             0|             0|
|2018-06-10T14:45:24-04:00|         57104|             0|             0|             0|             0|             0|             0|
+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+--------------+

--- Latencies Summary (ms) ---
+-------------------------+----------+----------+----------+
| Date/Time               | reads    | writes   | commands |
|-------------------------|----------|----------|----------|
|2018-06-10T14:41:27-04:00|         6|       144|         0|
|2018-06-10T14:42:27-04:00|         0|         0|         0|
|2018-06-10T14:43:27-04:00|         0|         0|         0|
|2018-06-10T14:44:27-04:00|         0|         0|         0|
|2018-06-10T14:45:24-04:00|         1|      1149|         0|
+-------------------------+----------+----------+----------+

--- Metrics ---
+-------------------------+----------+------------+------------+--------------+----------+----------+----------+----------+
| Date/Time               | Scanned  | ScannedObj |ScanAndOrder|WriteConflicts| Deleted  | Inserted | Returned | Updated  |
|-------------------------|----------|------------|------------|--------------|----------|----------|----------|----------|
|2018-06-10T14:41:27-04:00|     13902|     1171360|           0|             0|         0|   1155141|   1168446|       576|
|2018-06-10T14:42:27-04:00|   2187951|     2511693|           0|             0|         0|     60139|   1443347|     60142|
|2018-06-10T14:43:27-04:00|   2134123|     2449339|           0|             0|         0|     58609|   1406597|     58603|
|2018-06-10T14:44:27-04:00|   2118987|     2431159|           0|             0|         0|     57971|   1391426|     57979|
|2018-06-10T14:45:24-04:00|    687064|      813110|           0|        948468|    125520|         0|    126046|         0|
+-------------------------+----------+------------+------------+--------------+----------+----------+----------+----------+

--- WiredTiger Cache Summary ---
+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+--------------+
|                         | MaxBytes     | Currently    | Tracked      | Modified     | Unmodified   | PagesRead    | PagesWritten |
| Date/Time               | Configured   | InCache      | DirtyBytes   | PagesEvicted | PagesEvicted | IntoCache    | FromCache    |
|-------------------------|--------------|--------------|--------------|--------------|--------------|--------------|--------------|
|2018-06-10T14:41:27-04:00|    1073741824|     218019713|      48171209|       6408346|       3295790|          4029|        305615|
|2018-06-10T14:42:27-04:00|    1073741824|     447487489|      43764032|       6412384|       3295904|           199|         12129|
|2018-06-10T14:43:27-04:00|    1073741824|     620387947|      52227558|       6417696|       3298318|            79|         12060|
|2018-06-10T14:44:27-04:00|    1073741824|     719461300|      38703625|       6422466|       3302759|           141|         12605|
|2018-06-10T14:45:24-04:00|    1073741824|     702624705|      95899555|       7103422|       3316866|         51929|         49306|
+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+--------------+

--- WiredTiger Concurrent Transactions Summary ---
+-------------------------+--------------------------------------------+--------------------------------------------+
|                         | Read Ticket                                | Write Ticket                               |
| Date/Time               | Available    | Out          | Total        | Available    | Out          | Total        |
|-------------------------|--------------|--------------|--------------|--------------|--------------|--------------|
|2018-06-10T14:40:25-04:00|           127|             1|           128|           126|             2|           128|
|2018-06-10T14:41:27-04:00|           127|             1|           128|           128|             0|           128|
|2018-06-10T14:42:27-04:00|           127|             1|           128|           128|             0|           128|
|2018-06-10T14:43:27-04:00|           127|             1|           128|           120|             8|           128|
|2018-06-10T14:44:27-04:00|           127|             1|           128|           118|            10|           128|
|2018-06-10T14:45:24-04:00|           127|             1|           128|           118|            10|           128|
+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+
```
