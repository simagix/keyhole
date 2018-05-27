# Keyhole - Load Test
Load test a cluster/replica.  A default cycle last six minutes.

- Populate data in first minute
- Perform CRUD operations during the second and third minutes
- Burst test during the fourth and fifth minutes
- Perform CRUD ops in the last minute

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
2018/05/27 12:21:19 cleanup mongodb://localhost/?replicaSet=replset
2018/05/27 12:21:21 dropping collection _KEYHOLE_88800 keyhole
2018/05/27 12:21:21 dropping database _KEYHOLE_88800
Total TPS: 600 (tps) * 20 (conns) = 12000, duration = 6 (mins)
2018-05-27T12:21:22-04:00 res:     798, virt:    5852, faults:    34
2018-05-27T12:21:32-04:00 data:  376.5 ->  703.7, rate   32.0 MB/sec
2018-05-27T12:21:42-04:00 data:  703.7 ->  983.9, rate   27.6 MB/sec
2018-05-27T12:21:52-04:00 data:  983.9 -> 1259.8, rate   26.9 MB/sec
2018-05-27T12:22:02-04:00 data: 1259.8 -> 1529.8, rate   26.9 MB/sec
2018-05-27T12:22:22-04:00 res:     691, virt:    5725, faults:    34, i: 1134347, q: 2277751, u:       0, d:       0, iops:   56868
2018-05-27T12:22:12-04:00 data: 1529.8 -> 1821.4, rate   28.9 MB/sec
2018-05-27T12:22:22-04:00 data: 1821.4 -> 1877.7, rate    5.6 MB/sec
2018-05-27T12:23:22-04:00 res:    1207, virt:    6274, faults:    34, i:   46793, q:  115299, u:       0, d:       0, iops:    2701
2018-05-27T12:24:22-04:00 res:    1212, virt:    6281, faults:    34, i:      40, q:   30848, u:       0, d:       0, iops:     514
2018-05-27T12:25:22-04:00 res:    1213, virt:    6280, faults:    34, i:      40, q:   41088, u:       0, d:       0, iops:     685
2018-05-27T12:26:22-04:00 res:    1214, virt:    6282, faults:    34, i:      33, q:   41041, u:       0, d:       0, iops:     684

stats written to /var/folders/mv/q3097r9j5kxb59sg1btgf2s80000gp/T//keyhole_stats.27058-05-27T12-21-19
filename /var/folders/mv/q3097r9j5kxb59sg1btgf2s80000gp/T//keyhole_stats.27058-05-27T12-21-19

--- Analytic Summary ---
+-------------------------+-------+-------+------+--------+--------+--------+--------+--------+--------+--------+
| Date/Time               | res   | virt  | fault| Command| Delete | Getmore| Insert | Query  | Update | iops   |
|-------------------------|-------+-------|------|--------|--------|--------|--------|--------|--------|--------|
|2018-05-27T12:22:22-04:00|    691|   5725|    34|    2243|       7|    2918| 1138315|      14|       7|   19058|
|2018-05-27T12:23:22-04:00|   1207|   6274|    34|     493|      33|     414|   42825|     106|      33|     731|
|2018-05-27T12:24:22-04:00|   1212|   6281|    34|     323|      40|      93|      40|     101|      40|      10|
|2018-05-27T12:25:22-04:00|   1213|   6280|    34|     312|      40|      94|      40|     120|      40|      10|
|2018-05-27T12:26:22-04:00|   1214|   6282|    34|     280|      22|      64|      33|      85|      24|       8|
|2018-05-27T12:27:19-04:00|   1226|   6296|    34|     276|      38|      67|      27|     115|      36|       9|
+-------------------------+-------+-------+------+--------+--------+--------+--------+--------+--------+--------+

--- Latencies Summary ---
+-------------------------+----------+----------+----------+
| Date/Time               | reads    | writes   | commands |
|-------------------------|----------|----------|----------|
|2018-05-27T12:22:22-04:00|     15857|    221235|       370|
|2018-05-27T12:23:22-04:00|   2216731|    183311|       192|
|2018-05-27T12:24:22-04:00|   5170827|       219|        53|
|2018-05-27T12:25:22-04:00|   5893448|       185|        49|
|2018-05-27T12:26:22-04:00|   7746017|       124|        43|
|2018-05-27T12:27:19-04:00|   5964738|       106|        54|
+-------------------------+----------+----------+----------+

--- Metrics ---
+-------------------------+----------+------------+------------+--------------+----------+----------+----------+----------+
| Date/Time               | Scanned  | ScannedObj |ScanAndOrder|WriteConflicts| Deleted  | Inserted | Returned | Updated  |
|-------------------------|----------|------------|------------|--------------|----------|----------|----------|----------|
|2018-05-27T12:22:22-04:00|         0|     2277808|           0|             0|         0|   1134347|   2277751|         0|
|2018-05-27T12:23:22-04:00|         0|    70952400|          60|             0|         0|     46793|    115299|         0|
|2018-05-27T12:24:22-04:00|         0|    70869603|          60|             0|         0|        40|     30848|         0|
|2018-05-27T12:25:22-04:00|         0|    94495606|          80|             0|         0|        40|     41088|         0|
|2018-05-27T12:26:22-04:00|         0|    94498674|          80|             0|         0|        33|     41041|         0|
|2018-05-27T12:27:19-04:00|         0|    70876331|          60|             0|         0|        27|     30837|         0|
+-------------------------+----------+------------+------------+--------------+----------+----------+----------+----------+

--- WiredTiger Summary ---
+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+
|                         | MaxBytes     | Currently    | Unmodified   | Tracked      | PagesRead    | PagesWritten |
| Date/Time               | Configured   | InCache      | PagesEvicted | DirtyBytes   | IntoCache    | FromCache    |
|-------------------------|--------------|--------------|--------------|--------------|--------------|--------------|
|2018-05-27T12:22:22-04:00|    1073741824|     364296369|       1006214|     221501396|          3735|        303271|
|2018-05-27T12:23:22-04:00|    1073741824|     860014830|       1350819|        907786|        391863|         18459|
|2018-05-27T12:24:22-04:00|    1073741824|     872148628|       1792368|        746124|        448943|            33|
|2018-05-27T12:25:22-04:00|    1073741824|     935535711|       2234435|       1581141|        441703|            33|
|2018-05-27T12:26:22-04:00|    1073741824|     905879931|       2676562|        881337|        440495|            33|
|2018-05-27T12:27:19-04:00|    1073741824|     943669130|       3100164|        868057|        427968|            33|
+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+
2018/05/27 12:27:19 cleanup mongodb://localhost/?replicaSet=replset
2018/05/27 12:27:21 dropping collection _KEYHOLE_88800 keyhole
2018/05/27 12:27:21 dropping database _KEYHOLE_88800
```
