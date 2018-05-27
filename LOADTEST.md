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
2018/05/27 14:03:42 cleanup mongodb://localhost/?replicaSet=replset
2018/05/27 14:03:44 dropping collection _KEYHOLE_88800 keyhole
2018/05/27 14:03:44 dropping database _KEYHOLE_88800
Total TPS: 600 (tps) * 20 (conns) = 12000, duration = 6 (mins)
2018-05-27T14:03:44-04:00 Memory - resident:    1118, virtual:    6177
2018-05-27T14:03:54-04:00 Storage:  349.5 ->  668.8, rate   31.2 MB/sec
2018-05-27T14:04:04-04:00 Storage:  668.8 -> 1001.5, rate   33.1 MB/sec
2018-05-27T14:04:17-04:00 Storage: 1001.5 -> 1332.1, rate   26.8 MB/sec
2018-05-27T14:04:27-04:00 Storage: 1332.1 -> 1601.0, rate   26.6 MB/sec
2018-05-27T14:04:44-04:00 Memory - resident:     873, virtual:    5966, page faults:    34
2018-05-27T14:04:44-04:00 CRUD - i: 1130903, q: 2268755, u:       0, d:       0, iops:   56660
2018-05-27T14:04:44-04:00 Latency (ms) - read:    14.1, write:   219.7, command:     0.3
2018-05-27T14:04:37-04:00 Storage: 1601.0 -> 1878.4, rate   26.8 MB/sec
2018-05-27T14:04:47-04:00 Storage: 1878.4 -> 1906.9, rate    2.8 MB/sec
2018-05-27T14:05:44-04:00 Memory - resident:    1240, virtual:    6312, page faults:     0
2018-05-27T14:05:44-04:00 CRUD - i:   68837, q:  151253, u:       0, d:       0, iops:    3668
2018-05-27T14:05:44-04:00 Latency (ms) - read:  2176.1, write:   246.9, command:     0.1
2018-05-27T14:06:44-04:00 Memory - resident:    1228, virtual:    6296, page faults:     0
2018-05-27T14:06:44-04:00 CRUD - i:      40, q:   41090, u:       0, d:       0, iops:     685
2018-05-27T14:06:44-04:00 Latency (ms) - read:  6486.9, write:     0.2, command:     0.1
2018-05-27T14:07:44-04:00 Memory - resident:    1231, virtual:    6299, page faults:     0
2018-05-27T14:07:44-04:00 CRUD - i:      20, q:   30788, u:       0, d:       0, iops:     513
2018-05-27T14:07:44-04:00 Latency (ms) - read:  6801.7, write:     0.1, command:     0.0
2018-05-27T14:08:44-04:00 Memory - resident:    1236, virtual:    6304, page faults:     0
2018-05-27T14:08:44-04:00 CRUD - i:      60, q:   51388, u:       0, d:       0, iops:     857
2018-05-27T14:08:44-04:00 Latency (ms) - read:  4997.8, write:     0.2, command:     0.1

stats written to /var/folders/mv/q3097r9j5kxb59sg1btgf2s80000gp/T//keyhole_stats.27058-05-27T14-03-42
filename /var/folders/mv/q3097r9j5kxb59sg1btgf2s80000gp/T//keyhole_stats.27058-05-27T14-03-42

--- Analytic Summary ---
+-------------------------+-------+-------+------+--------+--------+--------+--------+--------+--------+--------+
| Date/Time               | res   | virt  | fault| Command| Delete | Getmore| Insert | Query  | Update | iops   |
|-------------------------|-------+-------|------|--------|--------|--------|--------|--------|--------|--------|
|2018-05-27T14:04:44-04:00|    873|   5966|    34|    2164|       3|    3098| 1134615|       6|       3|   18998|
|2018-05-27T14:05:44-04:00|   1240|   6312|    34|     475|      37|     319|   65125|      94|      37|    1101|
|2018-05-27T14:06:44-04:00|   1228|   6296|    34|     307|      40|      90|      40|     120|      40|      10|
|2018-05-27T14:07:44-04:00|   1231|   6299|    34|     274|      20|      60|      20|      80|      20|       7|
|2018-05-27T14:08:44-04:00|   1236|   6304|    34|     367|      60|     130|      60|     161|      60|      13|
|2018-05-27T14:09:42-04:00|   1228|   6295|    34|     379|      60|     160|      60|     195|      60|      16|
+-------------------------+-------+-------+------+--------+--------+--------+--------+--------+--------+--------+

--- Latencies Summary (ms) ---
+-------------------------+----------+----------+----------+
| Date/Time               | reads    | writes   | commands |
|-------------------------|----------|----------|----------|
|2018-05-27T14:04:44-04:00|        14|       219|         0|
|2018-05-27T14:05:44-04:00|      2176|       246|         0|
|2018-05-27T14:06:44-04:00|      6486|         0|         0|
|2018-05-27T14:07:44-04:00|      6801|         0|         0|
|2018-05-27T14:08:44-04:00|      4997|         0|         0|
|2018-05-27T14:09:42-04:00|      3271|         0|         0|
+-------------------------+----------+----------+----------+

--- Metrics ---
+-------------------------+----------+------------+------------+--------------+----------+----------+----------+----------+
| Date/Time               | Scanned  | ScannedObj |ScanAndOrder|WriteConflicts| Deleted  | Inserted | Returned | Updated  |
|-------------------------|----------|------------|------------|--------------|----------|----------|----------|----------|
|2018-05-27T14:04:44-04:00|         0|     2268775|           0|             0|         0|   1130903|   2268755|         0|
|2018-05-27T14:05:44-04:00|         0|    48119868|          40|             0|         0|     68837|    151253|         0|
|2018-05-27T14:06:44-04:00|         0|    95980424|          80|             0|         0|        40|     41090|         0|
|2018-05-27T14:07:44-04:00|         0|    71987423|          60|             0|         0|        20|     30788|         0|
|2018-05-27T14:08:44-04:00|         0|   119983023|         100|             0|         0|        60|     51388|         0|
|2018-05-27T14:09:42-04:00|         0|   164385216|         137|             0|         0|        60|     70324|         0|
+-------------------------+----------+------------+------------+--------------+----------+----------+----------+----------+

--- WiredTiger Summary ---
+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+
|                         | MaxBytes     | Currently    | Unmodified   | Tracked      | PagesRead    | PagesWritten |
| Date/Time               | Configured   | InCache      | PagesEvicted | DirtyBytes   | IntoCache    | FromCache    |
|-------------------------|--------------|--------------|--------------|--------------|--------------|--------------|
|2018-05-27T14:04:44-04:00|    1073741824|     344455309|      63969867|     222820536|          7575|        309904|
|2018-05-27T14:05:44-04:00|    1073741824|     937830738|      64292370|        703313|        379983|         17748|
|2018-05-27T14:06:44-04:00|    1073741824|     913112856|      64743850|       1249656|        448089|            32|
|2018-05-27T14:07:44-04:00|    1073741824|     857650584|      65185749|       1184243|        441533|            33|
|2018-05-27T14:08:44-04:00|    1073741824|     910184541|      65832856|       1458690|        642806|            32|
|2018-05-27T14:09:42-04:00|    1073741824|     915072379|      67603968|        979780|       1773856|            54|
+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+
2018/05/27 14:09:42 cleanup mongodb://localhost/?replicaSet=replset
2018/05/27 14:09:44 dropping collection _KEYHOLE_88800 keyhole
2018/05/27 14:09:44 dropping database _KEYHOLE_88800
```
