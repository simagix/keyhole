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
$ keyhole --uri mongodb://localhost/?replicaSet=replset --drop --cleanup

MongoDB URI: mongodb://localhost/?replicaSet=replset
Duration in minute(s): 5
2018/06/10 07:37:58 cleanup mongodb://localhost/?replicaSet=replset
2018/06/10 07:37:58 dropping collection _KEYHOLE_88800 keyhole
2018/06/10 07:37:58 dropping database _KEYHOLE_88800
Total TPS: 60 (tps) * 10 (conns) = 600, duration: 5 (mins), bulk size: 512
CollectServerStatus: connect to replset
CollectDBStats: connect to replset, _KEYHOLE_88800

2018-06-10T07:37:58-04:00 [replset] Memory - resident: 1410, virtual: 6428
2018-06-10T07:38:10-04:00 [replset] Storage: 365.9 -> 758.5, rate: 38.7 MB/sec
2018-06-10T07:38:20-04:00 [replset] Storage: 758.5 -> 1061.1, rate: 29.5 MB/sec
2018-06-10T07:38:30-04:00 [replset] Storage: 1061.1 -> 1339.7, rate: 27.7 MB/sec
2018-06-10T07:38:40-04:00 [replset] Storage: 1339.7 -> 1703.7, rate: 35.9 MB/sec
2018-06-10T07:38:50-04:00 [replset] Storage: 1703.7 -> 1978.7, rate: 27.2 MB/sec

2018-06-10T07:39:00-04:00 [replset] Memory - resident: 1107, virtual: 6125, page faults: 0, iops: 40844.9
2018-06-10T07:39:00-04:00 [replset] CRUD+  - insert: 1225226, find: 31, update: 10, delete: 0, getmore: 1807, command: 1972
2018-06-10T07:39:00-04:00 [replset] Latency- read: 12.9, write: 199.6, command: 0.5 (ms)
2018-06-10T07:39:00-04:00 [replset] Storage: 1978.7 -> 1979.7, rate: 0.1 MB/sec
2018-06-10T07:39:10-04:00 [replset] Storage: 1979.7 -> 1980.6, rate: 0.1 MB/sec
2018-06-10T07:39:20-04:00 [replset] Storage: 1980.6 -> 1981.6, rate: 0.1 MB/sec
CollectDBStats exiting...

2018-06-10T07:40:01-04:00 [replset] Memory - resident: 1097, virtual: 6120, page faults: 0, iops: 1541.8
2018-06-10T07:40:01-04:00 [replset] CRUD+  - insert: 3558, find: 10676, update: 3558, delete: 0, getmore: 1266, command: 1519
2018-06-10T07:40:01-04:00 [replset] Latency- read: 0.2, write: 0.2, command: 0.1 (ms)

2018-06-10T07:41:01-04:00 [replset] Memory - resident: 1089, virtual: 6116, page faults: 0, iops: 2716.2
2018-06-10T07:41:01-04:00 [replset] CRUD+  - insert: 6268, find: 18805, update: 6268, delete: 0, getmore: 1534, command: 1778
2018-06-10T07:41:01-04:00 [replset] Latency- read: 0.2, write: 0.2, command: 0.1 (ms)
thrashing average TPS was 55, lower than original 60
thrashing average TPS was 55, lower than original 60
thrashing average TPS was 55, lower than original 60
thrashing average TPS was 55, lower than original 60
thrashing average TPS was 55, lower than original 60
thrashing average TPS was 55, lower than original 60
thrashing average TPS was 55, lower than original 60
thrashing average TPS was 55, lower than original 60
thrashing average TPS was 55, lower than original 60
thrashing average TPS was 55, lower than original 60

2018-06-10T07:42:01-04:00 [replset] Memory - resident: 1087, virtual: 6119, page faults: 0, iops: 2717.0
2018-06-10T07:42:01-04:00 [replset] CRUD+  - insert: 6270, find: 18810, update: 6270, delete: 0, getmore: 1259, command: 1523
2018-06-10T07:42:01-04:00 [replset] Latency- read: 0.2, write: 0.2, command: 0.1 (ms)

stats written to /var/folders/mv/q3097r9j5kxb59sg1btgf2s80000gp/T//keyhole_stats.2018-06-10T073758-replset
--- Host: Kens-MBP, version: 3.6.4 ---

--- Analytic Summary ---
+-------------------------+-------+-------+------+--------+--------+--------+--------+--------+--------+--------+
| Date/Time               | res   | virt  | fault| Command| Delete | Getmore| Insert | Query  | Update | iops   |
|-------------------------|-------+-------|------|--------|--------|--------|--------|--------|--------|--------|
|2018-06-10T07:39:00-04:00|   1107|   6125|     0|    1972|       0|    1807| 1225226|      31|      10|   20148|
|2018-06-10T07:40:01-04:00|   1097|   6120|     0|    1519|       0|    1266|    3558|   10676|    3558|     342|
|2018-06-10T07:41:01-04:00|   1089|   6116|     0|    1778|       0|    1534|    6268|   18805|    6268|     577|
|2018-06-10T07:42:01-04:00|   1087|   6119|     0|    1523|       0|    1259|    6270|   18810|    6270|     568|
|2018-06-10T07:42:58-04:00|   1077|   6117|     0|    1687|    3420|    1479|     694|    2084|     694|     179|
+-------------------------+-------+-------+------+--------+--------+--------+--------+--------+--------+--------+

--- Global Locks Summary ---
+-------------------------+--------------+--------------------------------------------+--------------------------------------------+
|                         | Total Time   | Active Clients                             | Current Queue                              |
| Date/Time               | (ms)         | total        | readers      | writers      | total        | readers      | writers      |
|-------------------------|--------------|--------------|--------------|--------------|--------------|--------------|--------------|
|2018-06-10T07:39:00-04:00|         61976|             0|             0|             0|             0|             0|             0|
|2018-06-10T07:40:01-04:00|         60803|             0|             0|             0|             0|             0|             0|
|2018-06-10T07:41:01-04:00|         60013|             0|             0|             0|             0|             0|             0|
|2018-06-10T07:42:01-04:00|         60009|             0|             0|             0|             0|             0|             0|
|2018-06-10T07:42:58-04:00|         56804|             0|             0|             0|             0|             0|             0|
+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+--------------+

--- Latencies Summary (ms) ---
+-------------------------+----------+----------+----------+
| Date/Time               | reads    | writes   | commands |
|-------------------------|----------|----------|----------|
|2018-06-10T07:39:00-04:00|        12|       199|         0|
|2018-06-10T07:40:01-04:00|         0|         0|         0|
|2018-06-10T07:41:01-04:00|         0|         0|         0|
|2018-06-10T07:42:01-04:00|         0|         0|         0|
|2018-06-10T07:42:58-04:00|         0|        30|         0|
+-------------------------+----------+----------+----------+

--- Metrics ---
+-------------------------+----------+------------+------------+--------------+----------+----------+----------+----------+
| Date/Time               | Scanned  | ScannedObj |ScanAndOrder|WriteConflicts| Deleted  | Inserted | Returned | Updated  |
|-------------------------|----------|------------|------------|--------------|----------|----------|----------|----------|
|2018-06-10T07:39:00-04:00|       220|     1225467|           0|             0|         0|   1225226|   1225457|        10|
|2018-06-10T07:40:01-04:00|     78276|       91324|           0|             0|         0|      3558|     85394|      3558|
|2018-06-10T07:41:01-04:00|    139028|      163521|           0|             0|         0|      6268|    150433|      6268|
|2018-06-10T07:42:01-04:00|    139080|      163590|           0|             0|         0|      6270|    150480|      6270|
|2018-06-10T07:42:58-04:00|    275400|      321221|           0|        340728|     43123|       694|     59782|       694|
+-------------------------+----------+------------+------------+--------------+----------+----------+----------+----------+

--- WiredTiger Cache Summary ---
+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+--------------+
|                         | MaxBytes     | Currently    | Tracked      | Modified     | Unmodified   | PagesRead    | PagesWritten |
| Date/Time               | Configured   | InCache      | DirtyBytes   | PagesEvicted | PagesEvicted | IntoCache    | FromCache    |
|-------------------------|--------------|--------------|--------------|--------------|--------------|--------------|--------------|
|2018-06-10T07:39:00-04:00|    1073741824|     214565364|      54684722|       2932476|       1280828|          7959|        322117|
|2018-06-10T07:40:01-04:00|    1073741824|     227415426|      31995352|       2932582|       1280828|           106|          1455|
|2018-06-10T07:41:01-04:00|    1073741824|     268545720|      50324207|       2932665|       1280828|            97|           742|
|2018-06-10T07:42:01-04:00|    1073741824|     296331793|      34728017|       2932775|       1280828|            79|           968|
|2018-06-10T07:42:58-04:00|    1073741824|     269142067|      60112135|       2969716|       1281417|          8507|         32385|
+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+--------------+

--- WiredTiger Concurrent Transactions Summary ---
+-------------------------+--------------------------------------------+--------------------------------------------+
|                         | Read Ticket                                | Write Ticket                               |
| Date/Time               | Available    | Out          | Total        | Available    | Out          | Total        |
|-------------------------|--------------|--------------|--------------|--------------|--------------|--------------|
|2018-06-10T07:37:58-04:00|           127|             1|           128|           128|             0|           128|
|2018-06-10T07:39:00-04:00|           127|             1|           128|           128|             0|           128|
|2018-06-10T07:40:01-04:00|           127|             1|           128|           128|             0|           128|
|2018-06-10T07:41:01-04:00|           127|             1|           128|           128|             0|           128|
|2018-06-10T07:42:01-04:00|           127|             1|           128|           128|             0|           128|
|2018-06-10T07:42:58-04:00|           127|             1|           128|           128|             0|           128|
+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+
2018/06/10 07:42:58 cleanup mongodb://localhost/?replicaSet=replset
2018/06/10 07:42:58 dropping collection _KEYHOLE_88800 keyhole
2018/06/10 07:42:58 dropping database _KEYHOLE_88800
```
