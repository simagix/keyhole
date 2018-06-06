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

MongoDB URI: mongodb://user:password@localhost/
2018/06/04 11:22:14 cleanup mongodb://user:password@localhost/
2018/06/04 11:22:16 dropping collection _KEYHOLE_88800 keyhole
2018/06/04 11:22:16 dropping database _KEYHOLE_88800
Total TPS: 600 (tps) * 20 (conns) = 12000, duration = 5 (mins)

2018-06-04T11:22:16-04:00 Memory - resident: 810, virtual: 5734
2018-06-04T11:22:26-04:00 Storage: 922.0 -> 1741.2, rate: 81.4 MB/sec
2018-06-04T11:22:37-04:00 Storage: 1741.2 -> 2635.4, rate: 87.9 MB/sec
2018-06-04T11:22:47-04:00 Storage: 2635.4 -> 3512.0, rate: 87.0 MB/sec
2018-06-04T11:22:57-04:00 Storage: 3512.0 -> 4397.8, rate: 87.9 MB/sec

2018-06-04T11:23:16-04:00 Memory - resident: 1110, virtual: 6034, page faults: 0, iops: 54903.0
2018-06-04T11:23:16-04:00 CRUD+  - insert: 3294436, find: 0, update: 0, delete: 0, getmore: 0, command: 2155
2018-06-04T11:23:16-04:00 Latency- read: NaN, write: 6.7, command: 2.3 (ms)
2018-06-04T11:23:07-04:00 Storage: 4397.8 -> 5269.5, rate: 86.2 MB/sec
2018-06-04T11:23:17-04:00 Storage: 5269.5 -> 6112.5, rate: 82.7 MB/sec

2018-06-04T11:24:16-04:00 Memory - resident: 1141, virtual: 6079, page faults: 0, iops: 17973.8
2018-06-04T11:24:16-04:00 CRUD+  - insert: 620804, find: 369, update: 120, delete: 85, getmore: 0, command: 648
2018-06-04T11:24:16-04:00 Latency- read: 355.8, write: 255.6, command: 1.3 (ms)

2018-06-04T11:25:17-04:00 Memory - resident: 1149, virtual: 6085, page faults: 0, iops: 11671.3
2018-06-04T11:25:17-04:00 CRUD+  - insert: 0, find: 198, update: 65, delete: 270, getmore: 0, command: 199
2018-06-04T11:25:17-04:00 Latency- read: 722.3, write: 2885.5, command: 2.2 (ms)

2018-06-04T11:26:17-04:00 Memory - resident: 1148, virtual: 6085, page faults: 0, iops: 5068.9
2018-06-04T11:26:17-04:00 CRUD+  - insert: 0, find: 123, update: 35, delete: 186, getmore: 0, command: 279
2018-06-04T11:26:17-04:00 Latency- read: 3376.1, write: 3591.1, command: 2.5 (ms)

stats written to /var/folders/mv/q3097r9j5kxb59sg1btgf2s80000gp/T//keyhole_stats.2018-06-04T112214
filename /var/folders/mv/q3097r9j5kxb59sg1btgf2s80000gp/T//keyhole_stats.2018-06-04T112214

--- Analytic Summary ---
+-------------------------+-------+-------+------+--------+--------+--------+--------+--------+--------+--------+
| Date/Time               | res   | virt  | fault| Command| Delete | Getmore| Insert | Query  | Update | iops   |
|-------------------------|-------+-------|------|--------|--------|--------|--------|--------|--------|--------|
|2018-06-04T11:23:16-04:00|   1110|   6034|     0|    2155|       0|       0| 3294436|       0|       0|   54943|
|2018-06-04T11:24:16-04:00|   1141|   6079|     0|     648|      85|       0|  620804|     369|     120|   10367|
|2018-06-04T11:25:17-04:00|   1149|   6085|     0|     199|     270|       0|       0|     198|      65|      12|
|2018-06-04T11:26:17-04:00|   1148|   6085|     0|     279|     186|       0|       0|     123|      35|      10|
|2018-06-04T11:27:14-04:00|   1147|   6070|     0|     259|     307|       0|       0|      51|      11|      11|
+-------------------------+-------+-------+------+--------+--------+--------+--------+--------+--------+--------+

--- Latencies Summary (ms) ---
+-------------------------+----------+----------+----------+
| Date/Time               | reads    | writes   | commands |
|-------------------------|----------|----------|----------|
|2018-06-04T11:23:16-04:00|         0|         6|         2|
|2018-06-04T11:24:16-04:00|       355|       255|         1|
|2018-06-04T11:25:17-04:00|       722|      2885|         2|
|2018-06-04T11:26:17-04:00|      3376|      3591|         2|
|2018-06-04T11:27:14-04:00|      6681|      2366|         2|
+-------------------------+----------+----------+----------+

--- Metrics ---
+-------------------------+----------+------------+------------+--------------+----------+----------+----------+----------+
| Date/Time               | Scanned  | ScannedObj |ScanAndOrder|WriteConflicts| Deleted  | Inserted | Returned | Updated  |
|-------------------------|----------|------------|------------|--------------|----------|----------|----------|----------|
|2018-06-04T11:23:16-04:00|         0|           0|           0|             0|         0|   3294180|         0|         0|
|2018-06-04T11:24:16-04:00|  34184304|    34271988|           0|       1037348|    454680|    621060|      2688|         0|
|2018-06-04T11:25:17-04:00|  66089200|    66137045|           0|       1130095|    698860|         0|      1418|         0|
|2018-06-04T11:26:17-04:00|  91126906|    91161881|           0|        376428|    303120|         0|      1015|         0|
|2018-06-04T11:27:14-04:00|  80642500|    80654431|           0|             0|         0|         0|       491|         0|
+-------------------------+----------+------------+------------+--------------+----------+----------+----------+----------+

--- WiredTiger Cache Summary ---
+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+--------------+
|                         | MaxBytes     | Currently    | Tracked      | Modified     | Unmodified   | PagesRead    | PagesWritten |
| Date/Time               | Configured   | InCache      | DirtyBytes   | PagesEvicted | PagesEvicted | IntoCache    | FromCache    |
|-------------------------|--------------|--------------|--------------|--------------|--------------|--------------|--------------|
|2018-06-04T11:23:16-04:00|    1073741824|     887592125|      75742698|        531222|      11649554|           154|        201498|
|2018-06-04T11:24:16-04:00|    1073741824|     940653361|      49258739|        614437|      12535466|        906465|        118725|
|2018-06-04T11:25:17-04:00|    1073741824|     878593502|      56746728|        688629|      13554609|       1055344|         59576|
|2018-06-04T11:26:17-04:00|    1073741824|     931254283|      54606822|        724957|      14944294|       1421945|         18887|
|2018-06-04T11:27:14-04:00|    1073741824|     901122734|        382327|        726511|      16957247|       2005066|          2216|
+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+--------------+

--- WiredTiger Concurrent Transactions Summary ---
+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+
|                         | Read Ticket  | Read Ticket  | Read Ticket  | Write Ticket | Write Ticket | Write Ticket |
| Date/Time               | Available    | Out          | Total        | Available    | Out          | Total        |
|-------------------------|--------------|--------------|--------------|--------------|--------------|--------------|
|2018-06-04T11:22:16-04:00|           127|             1|           128|           127|             1|           128|
|2018-06-04T11:23:16-04:00|           127|             1|           128|           128|             0|           128|
|2018-06-04T11:24:16-04:00|           124|             4|           128|           120|             8|           128|
|2018-06-04T11:25:17-04:00|           125|             3|           128|           110|            18|           128|
|2018-06-04T11:26:17-04:00|           126|             2|           128|           114|            14|           128|
|2018-06-04T11:27:14-04:00|           123|             5|           128|           116|            12|           128|
+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+
2018/06/04 11:27:14 cleanup mongodb://user:password@localhost/
2018/06/04 11:27:16 dropping collection _KEYHOLE_88800 keyhole
2018/06/04 11:27:16 dropping database _KEYHOLE_88800
```
