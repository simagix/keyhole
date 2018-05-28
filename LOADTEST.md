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
2018/05/27 20:57:40 cleanup mongodb://localhost/?replicaSet=replset
2018/05/27 20:57:42 dropping collection _KEYHOLE_88800 keyhole
2018/05/27 20:57:42 dropping database _KEYHOLE_88800
Total TPS: 600 (tps) * 20 (conns) = 12000, duration = 5 (mins)

2018-05-27T20:57:42-04:00 Memory - resident:     853, virtual:    5936
2018-05-27T20:57:53-04:00 Storage:  317.1 ->  555.9, rate   23.7 MB/sec
2018-05-27T20:58:03-04:00 Storage:  555.9 ->  797.7, rate   24.0 MB/sec
2018-05-27T20:58:13-04:00 Storage:  797.7 -> 1064.3, rate   26.3 MB/sec
2018-05-27T20:58:23-04:00 Storage: 1064.3 -> 1321.2, rate   25.1 MB/sec

2018-05-27T20:58:42-04:00 Memory - resident:     717, virtual:    5790, page faults:   0, iops:   49580
2018-05-27T20:58:42-04:00 CRUD   - insert:  990020, find: 1984836, update:       3, delete:       0
2018-05-27T20:58:42-04:00 Latency- read:    14.6, write:   198.8, command:     5.2 (ms)
2018-05-27T20:58:33-04:00 Storage: 1321.2 -> 1583.1, rate   25.6 MB/sec
2018-05-27T20:58:44-04:00 Storage: 1583.1 -> 1759.2, rate   16.9 MB/sec

2018-05-27T20:59:42-04:00 Memory - resident:     705, virtual:    5783, page faults:   0, iops:   34704
2018-05-27T20:59:42-04:00 CRUD   - insert:  116680, find: 1962754, update:    2827, delete:       0
2018-05-27T20:59:42-04:00 Latency- read:     2.1, write:    27.9, command:     0.2 (ms)

2018-05-27T21:00:43-04:00 Memory - resident:     704, virtual:    5781, page faults:   0, iops:   33910
2018-05-27T21:00:43-04:00 CRUD   - insert:       0, find: 2031312, update:    3308, delete:       0
2018-05-27T21:00:43-04:00 Latency- read:     1.5, write:     0.3, command:     0.1 (ms)

2018-05-27T21:01:43-04:00 Memory - resident:     704, virtual:    5781, page faults:   0, iops:   33511
2018-05-27T21:01:43-04:00 CRUD   - insert:       0, find: 2007440, update:    3270, delete:       0
2018-05-27T21:01:43-04:00 Latency- read:     1.5, write:     0.3, command:     0.1 (ms)

stats written to /var/folders/mv/q3097r9j5kxb59sg1btgf2s80000gp/T//keyhole_stats.27058-05-27T20-57-40
filename /var/folders/mv/q3097r9j5kxb59sg1btgf2s80000gp/T//keyhole_stats.27058-05-27T20-57-40

--- Analytic Summary ---
+-------------------------+-------+-------+------+--------+--------+--------+--------+--------+--------+--------+
| Date/Time               | res   | virt  | fault| Command| Delete | Getmore| Insert | Query  | Update | iops   |
|-------------------------|-------+-------|------|--------|--------|--------|--------|--------|--------|--------|
|2018-05-27T20:58:42-04:00|    717|   5790|     0|    2538|       0|    2732|  992068|       6|       3|   16622|
|2018-05-27T20:59:42-04:00|    705|   5783|     0|    6768|       0|    6288|  114632|    5635|    2827|    2269|
|2018-05-27T21:00:43-04:00|    704|   5781|     0|    8240|       0|    7468|       0|    6616|    3308|     427|
|2018-05-27T21:01:43-04:00|    704|   5781|     0|    8207|       1|    7463|       0|    6544|    3270|     424|
|2018-05-27T21:02:40-04:00|    749|   5866|     0|    6422|      19|    7589|       0|     704|     344|     264|
+-------------------------+-------+-------+------+--------+--------+--------+--------+--------+--------+--------+

--- Latencies Summary (ms) ---
+-------------------------+----------+----------+----------+
| Date/Time               | reads    | writes   | commands |
|-------------------------|----------|----------|----------|
|2018-05-27T20:58:42-04:00|        14|       198|         5|
|2018-05-27T20:59:42-04:00|         2|        27|         0|
|2018-05-27T21:00:43-04:00|         1|         0|         0|
|2018-05-27T21:01:43-04:00|         1|         0|         0|
|2018-05-27T21:02:40-04:00|         1|         0|         0|
+-------------------------+----------+----------+----------+

--- Metrics ---
+-------------------------+----------+------------+------------+--------------+----------+----------+----------+----------+
| Date/Time               | Scanned  | ScannedObj |ScanAndOrder|WriteConflicts| Deleted  | Inserted | Returned | Updated  |
|-------------------------|----------|------------|------------|--------------|----------|----------|----------|----------|
|2018-05-27T20:58:42-04:00|      1539|     1985530|           0|             0|         0|    990020|   1984836|         3|
|2018-05-27T20:59:42-04:00|   1448715|     2651637|           0|            29|         0|    116680|   1962754|      2827|
|2018-05-27T21:00:43-04:00|   1697516|     2844259|           0|             8|         0|         0|   2031312|      3308|
|2018-05-27T21:01:43-04:00|   1676486|     2810624|           0|             3|         0|         0|   2007440|      3270|
|2018-05-27T21:02:40-04:00|    178520|      447240|           0|             1|         0|         0|    360526|       344|
+-------------------------+----------+------------+------------+--------------+----------+----------+----------+----------+

--- WiredTiger Summary ---
+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+
|                         | MaxBytes     | Currently    | Unmodified   | Tracked      | PagesRead    | PagesWritten |
| Date/Time               | Configured   | InCache      | PagesEvicted | DirtyBytes   | IntoCache    | FromCache    |
|-------------------------|--------------|--------------|--------------|--------------|--------------|--------------|
|2018-05-27T20:58:42-04:00|    1073741824|     396254446|         47267|     222499083|          3009|        229466|
|2018-05-27T20:59:42-04:00|    1073741824|     299307459|         47271|        880128|           102|         32093|
|2018-05-27T21:00:43-04:00|    1073741824|     300215970|         47271|       1772085|            60|            57|
|2018-05-27T21:01:43-04:00|    1073741824|     312892603|         47271|       9297101|           432|            55|
|2018-05-27T21:02:40-04:00|    1073741824|     668099849|         47955|      72291551|         31682|         25414|
+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+
2018/05/27 21:02:40 cleanup mongodb://localhost/?replicaSet=replset
2018/05/27 21:02:42 dropping collection _KEYHOLE_88800 keyhole
2018/05/27 21:02:42 dropping database _KEYHOLE_88800
```
