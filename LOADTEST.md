# Keyhole - Load Test
Load test a cluster/replica.  A default cycle last six minutes.

- Populate data in first minute
- Perform CRUD operations during the second minutes
- Burst test until before the last minute
- Perform cleanup ops in the last minute

## Load Test
Keyhole uses a template to read in a sample document and then generates randomized documents based on the template (see [SEED.md](SEED,md) for details.  To execute a load test with a template, do

```
keyhole -uri=mongodb://localhost/?replicaSet=replset --file example/seedkeys.json
```

To execute a demo load test, run it without the `--file` flag.

```
keyhole -uri=mongodb://localhost/?replicaSet=replset
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
$ keyhole -uri=mongodb://localhost/?replicaSet=replset

MongoDB URI: mongodb://localhost/?replicaSet=replset
Duration in minute(s): 5
2018/06/09 17:01:25 cleanup mongodb://localhost/?replicaSet=replset
2018/06/09 17:01:27 dropping collection _KEYHOLE_88800 keyhole
2018/06/09 17:01:27 dropping database _KEYHOLE_88800
Total TPS: 600 (tps) * 20 (conns) = 12000, duration: 5 (mins), bulk size: 10
CollectServerStatus: connect to replset
CollectDBStats: connect to replset, _KEYHOLE_88800

2018-06-09T17:01:28-04:00 [replset] Memory - resident: 675, virtual: 5713
2018-06-09T17:01:39-04:00 [replset] Storage: 420.3 -> 770.1, rate: 34.8 MB/sec
2018-06-09T17:01:49-04:00 [replset] Storage: 770.1 -> 1106.3, rate: 33.3 MB/sec
2018-06-09T17:01:59-04:00 [replset] Storage: 1106.3 -> 1422.4, rate: 30.8 MB/sec
2018-06-09T17:02:09-04:00 [replset] Storage: 1422.4 -> 1761.6, rate: 33.8 MB/sec
2018-06-09T17:02:19-04:00 [replset] Storage: 1761.6 -> 2015.9, rate: 25.2 MB/sec

2018-06-09T17:02:30-04:00 [replset] Memory - resident: 786, virtual: 5837, page faults: 0, iops: 41708.1
2018-06-09T17:02:30-04:00 [replset] CRUD+  - insert: 1251984, find: 1, update: 0, delete: 0, getmore: 1899, command: 1740
2018-06-09T17:02:30-04:00 [replset] Latency- read: 13.2, write: 226.4, command: 0.3 (ms)
2018-06-09T17:02:29-04:00 [replset] Storage: 2015.9 -> 2308.5, rate: 29.2 MB/sec
2018-06-09T17:02:40-04:00 [replset] Storage: 2308.5 -> 2374.8, rate: 6.5 MB/sec
2018-06-09T17:02:50-04:00 [replset] Storage: 2374.8 -> 2391.8, rate: 1.7 MB/sec
2018-06-09T17:03:00-04:00 [replset] Storage: 2391.8 -> 2409.4, rate: 1.7 MB/sec
2018-06-09T17:03:10-04:00 [replset] Storage: 2409.4 -> 2426.4, rate: 1.7 MB/sec
2018-06-09T17:03:20-04:00 [replset] Storage: 2426.4 -> 2442.7, rate: 1.6 MB/sec

2018-06-09T17:03:30-04:00 [replset] Memory - resident: 760, virtual: 5804, page faults: 0, iops: 26758.1
2018-06-09T17:03:30-04:00 [replset] CRUD+  - insert: 260429, find: 147584, update: 49186, delete: 0, getmore: 4550, command: 7363
2018-06-09T17:03:30-04:00 [replset] Latency- read: 0.2, write: 2.0, command: 0.1 (ms)
2018-06-09T17:03:30-04:00 [replset] Storage: 2442.7 -> 2459.4, rate: 1.7 MB/sec
2018-06-09T17:03:40-04:00 [replset] Storage: 2459.4 -> 2476.2, rate: 1.7 MB/sec
2018-06-09T17:03:50-04:00 [replset] Storage: 2476.2 -> 2493.1, rate: 1.7 MB/sec
2018-06-09T17:04:00-04:00 [replset] Storage: 2493.1 -> 2509.8, rate: 1.7 MB/sec
2018-06-09T17:04:10-04:00 [replset] Storage: 2509.8 -> 2526.4, rate: 1.7 MB/sec
2018-06-09T17:04:20-04:00 [replset] Storage: 2526.4 -> 2543.2, rate: 1.7 MB/sec

2018-06-09T17:04:31-04:00 [replset] Memory - resident: 764, virtual: 5803, page faults: 0, iops: 24168.4
2018-06-09T17:04:31-04:00 [replset] CRUD+  - insert: 60419, find: 181266, update: 60428, delete: 0, getmore: 6721, command: 10728
2018-06-09T17:04:31-04:00 [replset] Latency- read: 0.2, write: 0.2, command: 0.1 (ms)
2018-06-09T17:04:30-04:00 [replset] Storage: 2543.2 -> 2559.8, rate: 1.7 MB/sec
2018-06-09T17:04:40-04:00 [replset] Storage: 2559.8 -> 2576.3, rate: 1.6 MB/sec
2018-06-09T17:04:50-04:00 [replset] Storage: 2576.3 -> 2593.0, rate: 1.7 MB/sec
2018-06-09T17:05:00-04:00 [replset] Storage: 2593.0 -> 2609.4, rate: 1.6 MB/sec
2018-06-09T17:05:10-04:00 [replset] Storage: 2609.4 -> 2626.3, rate: 1.7 MB/sec
2018-06-09T17:05:20-04:00 [replset] Storage: 2626.3 -> 2642.9, rate: 1.7 MB/sec

2018-06-09T17:05:31-04:00 [replset] Memory - resident: 793, virtual: 5846, page faults: 0, iops: 23887.8
2018-06-09T17:05:31-04:00 [replset] CRUD+  - insert: 59719, find: 179154, update: 59716, delete: 0, getmore: 6741, command: 10733
2018-06-09T17:05:31-04:00 [replset] Latency- read: 0.2, write: 0.1, command: 0.1 (ms)
2018-06-09T17:05:30-04:00 [replset] Storage: 2642.9 -> 2659.3, rate: 1.6 MB/sec
2018-06-09T17:05:40-04:00 [replset] Storage: 2659.3 -> 2675.4, rate: 1.6 MB/sec
2018-06-09T17:05:50-04:00 [replset] Storage: 2675.4 -> 2691.9, rate: 1.6 MB/sec
2018-06-09T17:06:00-04:00 [replset] Storage: 2691.9 -> 2708.5, rate: 1.7 MB/sec
2018-06-09T17:06:10-04:00 [replset] Storage: 2708.5 -> 2724.5, rate: 1.6 MB/sec

stats written to /var/folders/mv/q3097r9j5kxb59sg1btgf2s80000gp/T//keyhole_stats.2018-06-09T170125-replset
--- Host: Kens-MBP, version: 3.6.4 ---

--- Analytic Summary ---
+-------------------------+-------+-------+------+--------+--------+--------+--------+--------+--------+--------+
| Date/Time               | res   | virt  | fault| Command| Delete | Getmore| Insert | Query  | Update | iops   |
|-------------------------|-------+-------|------|--------|--------|--------|--------|--------|--------|--------|
|2018-06-09T17:02:30-04:00|    786|   5837|     0|    1740|       0|    1899| 1251984|       1|       0|   20584|
|2018-06-09T17:03:30-04:00|    760|   5804|     0|    7363|       0|    4550|  260429|  147584|   49186|    7818|
|2018-06-09T17:04:31-04:00|    764|   5803|     0|   10728|       0|    6721|   60419|  181266|   60428|    5326|
|2018-06-09T17:05:31-04:00|    793|   5846|     0|   10733|       0|    6741|   59719|  179154|   59716|    5267|
|2018-06-09T17:06:25-04:00|    871|   5947|     0|    9880|       0|    6086|   53042|  159121|   53037|    5206|
+-------------------------+-------+-------+------+--------+--------+--------+--------+--------+--------+--------+

--- Global Locks Summary ---
+-------------------------+--------------+--------------------------------------------+--------------------------------------------+
|                         | Total Time   | Active Clients                             | Current Queue                              |
| Date/Time               | (ms)         | total        | readers      | writers      | total        | readers      | writers      |
|-------------------------|--------------|--------------|--------------|--------------|--------------|--------------|--------------|
|2018-06-09T17:02:30-04:00|         61949|             0|             0|             0|             0|             0|             0|
|2018-06-09T17:03:30-04:00|         60681|             0|             0|             0|             0|             0|             0|
|2018-06-09T17:04:31-04:00|         60405|             0|             0|             0|             0|             0|             0|
|2018-06-09T17:05:31-04:00|         60152|             0|             0|             0|             0|             0|             0|
|2018-06-09T17:06:25-04:00|         54328|             0|             0|             0|             0|             0|             0|
+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+--------------+

--- Latencies Summary (ms) ---
+-------------------------+----------+----------+----------+
| Date/Time               | reads    | writes   | commands |
|-------------------------|----------|----------|----------|
|2018-06-09T17:02:30-04:00|        13|       226|         0|
|2018-06-09T17:03:30-04:00|         0|         2|         0|
|2018-06-09T17:04:31-04:00|         0|         0|         0|
|2018-06-09T17:05:31-04:00|         0|         0|         0|
|2018-06-09T17:06:25-04:00|         0|         0|         0|
+-------------------------+----------+----------+----------+

--- Metrics ---
+-------------------------+----------+------------+------------+--------------+----------+----------+----------+----------+
| Date/Time               | Scanned  | ScannedObj |ScanAndOrder|WriteConflicts| Deleted  | Inserted | Returned | Updated  |
|-------------------------|----------|------------|------------|--------------|----------|----------|----------|----------|
|2018-06-09T17:02:30-04:00|         0|     1253765|           0|             0|         0|   1248720|   1253765|         0|
|2018-06-09T17:03:30-04:00|   1687292|     2076888|           0|             0|         0|    263693|   1341793|         0|
|2018-06-09T17:04:31-04:00|   2085356|     2306151|           0|             0|         0|     60421|   1389681|         0|
|2018-06-09T17:05:31-04:00|   2060921|     2279137|           0|             0|         0|     59717|   1373553|         0|
|2018-06-09T17:06:25-04:00|   1818628|     2012638|           0|             0|         0|     53042|   1219948|         0|
+-------------------------+----------+------------+------------+--------------+----------+----------+----------+----------+

--- WiredTiger Cache Summary ---
+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+--------------+
|                         | MaxBytes     | Currently    | Tracked      | Modified     | Unmodified   | PagesRead    | PagesWritten |
| Date/Time               | Configured   | InCache      | DirtyBytes   | PagesEvicted | PagesEvicted | IntoCache    | FromCache    |
|-------------------------|--------------|--------------|--------------|--------------|--------------|--------------|--------------|
|2018-06-09T17:02:30-04:00|    1073741824|     252181057|     220048318|       2301725|         57796|          8996|        396763|
|2018-06-09T17:03:30-04:00|    1073741824|     297317642|      47346002|       2373424|         57913|           350|         82427|
|2018-06-09T17:04:31-04:00|    1073741824|     473574716|      50004594|       2373955|         60090|            95|          7840|
|2018-06-09T17:05:31-04:00|    1073741824|     593802590|      32193999|       2374953|         64023|            62|          8758|
|2018-06-09T17:06:25-04:00|    1073741824|     709610890|      56763528|       2375463|         67306|            62|          6512|
+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+--------------+

--- WiredTiger Concurrent Transactions Summary ---
+-------------------------+--------------------------------------------+--------------------------------------------+
|                         | Read Ticket                                | Write Ticket                               |
| Date/Time               | Available    | Out          | Total        | Available    | Out          | Total        |
|-------------------------|--------------|--------------|--------------|--------------|--------------|--------------|
|2018-06-09T17:01:28-04:00|           127|             1|           128|           127|             1|           128|
|2018-06-09T17:02:30-04:00|           126|             2|           128|           108|            20|           128|
|2018-06-09T17:03:30-04:00|           126|             2|           128|           128|             0|           128|
|2018-06-09T17:04:31-04:00|           127|             1|           128|           128|             0|           128|
|2018-06-09T17:05:31-04:00|           127|             1|           128|           128|             0|           128|
|2018-06-09T17:06:25-04:00|           126|             2|           128|           128|             0|           128|
+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+
2018/06/09 17:06:25 cleanup mongodb://localhost/?replicaSet=replset
2018/06/09 17:06:27 dropping collection _KEYHOLE_88800 keyhole
2018/06/09 17:06:27 dropping database _KEYHOLE_88800
```
