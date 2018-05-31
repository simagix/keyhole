# Keyhole - Load Test
Load test a cluster/replica.  A default cycle last six minutes.

- Populate data in first minute
- Perform CRUD operations during the second minutes
- Burst test until before the last minute
- Perform cleanup ops in the last minute

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
2018/05/31 05:47:04 cleanup mongodb://localhost/?replicaSet=replset
2018/05/31 05:47:06 dropping collection _KEYHOLE_88800 keyhole
2018/05/31 05:47:06 dropping database _KEYHOLE_88800
Total TPS: 600 (tps) * 20 (conns) = 12000, duration = 5 (mins)

2018-05-31T05:47:07-04:00 Memory - resident:    1000, virtual:    6011
2018-05-31T05:47:17-04:00 Storage:  381.9 ->  759.0, rate   36.8 MB/sec
2018-05-31T05:47:27-04:00 Storage:  759.0 -> 1102.2, rate   34.2 MB/sec
2018-05-31T05:47:37-04:00 Storage: 1102.2 -> 1481.4, rate   37.6 MB/sec
2018-05-31T05:47:47-04:00 Storage: 1481.4 -> 1775.4, rate   28.9 MB/sec

2018-05-31T05:48:07-04:00 Memory - resident:     955, virtual:    5972, page faults:   0, iops: 43096.9
2018-05-31T05:48:07-04:00 CRUD+  - insert:1291828, find:     10, update:      3, delete:      5, getmore:   1992, command:   2061
2018-05-31T05:48:07-04:00 Latency- read:    13.7, write:   214.7, command:     0.3 (ms)
2018-05-31T05:47:57-04:00 Storage: 1775.4 -> 2068.6, rate   28.9 MB/sec
2018-05-31T05:48:08-04:00 Storage: 2068.6 -> 2118.8, rate    4.9 MB/sec

2018-05-31T05:49:07-04:00 Memory - resident:    1140, virtual:    6190, page faults:   0, iops:  3285.4
2018-05-31T05:49:07-04:00 CRUD+  - insert:  43548, find:     16, update:      3, delete:     18, getmore:   3537, command:   1983
2018-05-31T05:49:07-04:00 Latency- read:    35.3, write:   136.5, command:     0.4 (ms)

2018-05-31T05:50:07-04:00 Memory - resident:    1204, virtual:    6276, page faults:   0, iops:  6154.1
2018-05-31T05:50:07-04:00 CRUD+  - insert:      0, find:     19, update:      3, delete:     24, getmore:   2777, command:   1582
2018-05-31T05:50:07-04:00 Latency- read:    67.2, write: 60228.6, command:    12.5 (ms)

2018-05-31T05:51:07-04:00 Memory - resident:    1208, virtual:    6267, page faults:   0, iops:  9904.3
2018-05-31T05:51:07-04:00 CRUD+  - insert:      0, find:     33, update:      6, delete:     38, getmore:   2673, command:   1450
2018-05-31T05:51:07-04:00 Latency- read:    54.9, write: 48354.1, command:     0.4 (ms)

stats written to /var/folders/mv/q3097r9j5kxb59sg1btgf2s80000gp/T//keyhole_stats.31058-05-31T05-47-04
filename /var/folders/mv/q3097r9j5kxb59sg1btgf2s80000gp/T//keyhole_stats.31058-05-31T05-47-04

--- Analytic Summary ---
+-------------------------+-------+-------+------+--------+--------+--------+--------+--------+--------+--------+
| Date/Time               | res   | virt  | fault| Command| Delete | Getmore| Insert | Query  | Update | iops   |
|-------------------------|-------+-------|------|--------|--------|--------|--------|--------|--------|--------|
|2018-05-31T05:48:07-04:00|    955|   5972|     0|    2061|       5|    1992| 1291828|      10|       3|   21598|
|2018-05-31T05:49:07-04:00|   1140|   6190|     0|    1983|      18|    3537|   43548|      16|       3|     818|
|2018-05-31T05:50:07-04:00|   1204|   6276|     0|    1582|      24|    2777|       0|      19|       3|      73|
|2018-05-31T05:51:07-04:00|   1208|   6267|     0|    1450|      38|    2673|       0|      33|       6|      70|
|2018-05-31T05:52:04-04:00|   1192|   6256|     0|    1242|     126|    2317|       0|       3|       0|      64|
+-------------------------+-------+-------+------+--------+--------+--------+--------+--------+--------+--------+

--- Latencies Summary (ms) ---
+-------------------------+----------+----------+----------+
| Date/Time               | reads    | writes   | commands |
|-------------------------|----------|----------|----------|
|2018-05-31T05:48:07-04:00|        13|       214|         0|
|2018-05-31T05:49:07-04:00|        35|       136|         0|
|2018-05-31T05:50:07-04:00|        67|     60228|        12|
|2018-05-31T05:51:07-04:00|        54|     48354|         0|
|2018-05-31T05:52:04-04:00|        13|     12710|         0|
+-------------------------+----------+----------+----------+

--- Metrics ---
+-------------------------+----------+------------+------------+--------------+----------+----------+----------+----------+
| Date/Time               | Scanned  | ScannedObj |ScanAndOrder|WriteConflicts| Deleted  | Inserted | Returned | Updated  |
|-------------------------|----------|------------|------------|--------------|----------|----------|----------|----------|
|2018-05-31T05:48:07-04:00|      1813|     1295348|           0|             0|         0|   1290740|   1295074|         0|
|2018-05-31T05:49:07-04:00|    404894|      556547|           0|             0|         0|     44700|    152424|         0|
|2018-05-31T05:50:07-04:00|   2462620|     2679646|           0|         29187|    151976|         0|    217271|         0|
|2018-05-31T05:51:07-04:00|   4332919|     4594013|           0|         46813|    331810|         0|    262446|         0|
|2018-05-31T05:52:04-04:00|   6307418|     6721487|           0|         83858|    481749|         0|    413920|         0|
+-------------------------+----------+------------+------------+--------------+----------+----------+----------+----------+

--- WiredTiger Summary ---
+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+
|                         | MaxBytes     | Currently    | Unmodified   | Tracked      | PagesRead    | PagesWritten |
| Date/Time               | Configured   | InCache      | PagesEvicted | DirtyBytes   | IntoCache    | FromCache    |
|-------------------------|--------------|--------------|--------------|--------------|--------------|--------------|
|2018-05-31T05:48:07-04:00|    1073741824|     328340043|       1451951|     219599174|          9446|        382065|
|2018-05-31T05:49:07-04:00|    1073741824|     870097724|       1586224|     212882407|        234378|         74166|
|2018-05-31T05:50:07-04:00|    1073741824|    1020059017|       1816179|     212484570|        420775|        170641|
|2018-05-31T05:51:07-04:00|    1073741824|    1020092431|       2381773|     149440632|        881341|        339776|
|2018-05-31T05:52:04-04:00|    1073741824|     776058432|       2945998|     224926799|       1013973|        513941|
+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+
2018/05/31 05:52:05 cleanup mongodb://localhost/?replicaSet=replset
2018/05/31 05:52:07 dropping collection _KEYHOLE_88800 keyhole
2018/05/31 05:52:07 dropping database _KEYHOLE_88800
```
