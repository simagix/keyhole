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
2018/05/27 13:50:47 cleanup mongodb://localhost/?replicaSet=replset
2018/05/27 13:50:49 dropping collection _KEYHOLE_88800 keyhole
2018/05/27 13:50:49 dropping database _KEYHOLE_88800
Total TPS: 600 (tps) * 20 (conns) = 12000, duration = 6 (mins)
2018-05-27T13:50:49-04:00 Memory - resident:    1179, virtual:    6236, page faults:    34
2018-05-27T13:51:00-04:00 Storage:  396.1 ->  670.1, rate   26.3 MB/sec
2018-05-27T13:51:10-04:00 Storage:  670.1 ->  934.5, rate   26.2 MB/sec
2018-05-27T13:51:20-04:00 Storage:  934.5 -> 1201.6, rate   26.4 MB/sec
2018-05-27T13:51:30-04:00 Storage: 1201.6 -> 1487.3, rate   27.5 MB/sec
2018-05-27T13:51:49-04:00 Memory - resident:     919, virtual:    5965, page faults:    34
2018-05-27T13:51:49-04:00 CRUD - i: 1068602, q: 2146192, u:       0, d:       0, iops:   53579
2018-05-27T13:51:49-04:00 Latency (ms) - read:    18.0, write:   240.3, command:     0.4
2018-05-27T13:51:40-04:00 Storage: 1487.3 -> 1761.1, rate   27.2 MB/sec
2018-05-27T13:51:51-04:00 Storage: 1761.1 -> 1832.9, rate    7.1 MB/sec
2018-05-27T13:52:49-04:00 Memory - resident:    1249, virtual:    6326, page faults:    34
2018-05-27T13:52:49-04:00 CRUD - i:   84638, q:  191058, u:       0, d:       0, iops:    4594
2018-05-27T13:52:49-04:00 Latency (ms) - read:  1726.3, write:   167.5, command:     0.1
2018-05-27T13:53:49-04:00 Memory - resident:    1232, virtual:    6309, page faults:    34
2018-05-27T13:53:49-04:00 CRUD - i:      60, q:   51386, u:       0, d:       0, iops:     857
2018-05-27T13:53:49-04:00 Latency (ms) - read:  3819.7, write:     0.2, command:     0.1
2018-05-27T13:54:49-04:00 Memory - resident:    1233, virtual:    6311, page faults:    34
2018-05-27T13:54:49-04:00 CRUD - i:      60, q:   70838, u:       0, d:       0, iops:    1181
2018-05-27T13:54:49-04:00 Latency (ms) - read:  3213.0, write:     0.1, command:     0.1
2018-05-27T13:55:49-04:00 Memory - resident:    1233, virtual:    6310, page faults:    34
2018-05-27T13:55:49-04:00 CRUD - i:      65, q:   64197, u:       0, d:       0, iops:    1071
2018-05-27T13:55:49-04:00 Latency (ms) - read:  3096.7, write:     0.1, command:     0.1

stats written to /var/folders/mv/q3097r9j5kxb59sg1btgf2s80000gp/T//keyhole_stats.27058-05-27T13-50-47
filename /var/folders/mv/q3097r9j5kxb59sg1btgf2s80000gp/T//keyhole_stats.27058-05-27T13-50-47

--- Analytic Summary ---
+-------------------------+-------+-------+------+--------+--------+--------+--------+--------+--------+--------+
| Date/Time               | res   | virt  | fault| Command| Delete | Getmore| Insert | Query  | Update | iops   |
|-------------------------|-------+-------|------|--------|--------|--------|--------|--------|--------|--------|
|2018-05-27T13:51:49-04:00|    919|   5965|    34|    2125|       2|    2692| 1072442|       4|       2|   17954|
|2018-05-27T13:52:49-04:00|   1249|   6326|    34|     470|      38|     475|   80798|     116|      38|    1365|
|2018-05-27T13:53:49-04:00|   1232|   6309|    34|     386|      60|     158|      60|     161|      60|      14|
|2018-05-27T13:54:49-04:00|   1233|   6311|    34|     385|      60|     168|      60|     198|      60|      15|
|2018-05-27T13:55:49-04:00|   1233|   6310|    34|     412|      65|     198|      65|     190|      65|      16|
|2018-05-27T13:56:47-04:00|   1232|   6309|    34|     368|      58|     144|      58|     172|      58|      15|
+-------------------------+-------+-------+------+--------+--------+--------+--------+--------+--------+--------+

--- Latencies Summary (ms) ---
+-------------------------+----------+----------+----------+
| Date/Time               | reads    | writes   | commands |
|-------------------------|----------|----------|----------|
|2018-05-27T13:51:49-04:00|        17|       240|         0|
|2018-05-27T13:52:49-04:00|      1726|       167|         0|
|2018-05-27T13:53:49-04:00|      3819|         0|         0|
|2018-05-27T13:54:49-04:00|      3213|         0|         0|
|2018-05-27T13:55:49-04:00|      3096|         0|         0|
|2018-05-27T13:56:47-04:00|      3490|         0|         0|
+-------------------------+----------+----------+----------+

--- Metrics ---
+-------------------------+----------+------------+------------+--------------+----------+----------+----------+----------+
| Date/Time               | Scanned  | ScannedObj |ScanAndOrder|WriteConflicts| Deleted  | Inserted | Returned | Updated  |
|-------------------------|----------|------------|------------|--------------|----------|----------|----------|----------|
|2018-05-27T13:51:49-04:00|         0|     2146228|           0|             0|         0|   1068602|   2146192|         0|
|2018-05-27T13:52:49-04:00|         0|    69354241|          60|             0|         0|     84638|    191058|         0|
|2018-05-27T13:53:49-04:00|         0|   115327091|         100|             0|         0|        60|     51386|         0|
|2018-05-27T13:54:49-04:00|         0|   159159874|         138|             0|         0|        60|     70838|         0|
|2018-05-27T13:55:49-04:00|         0|   144174921|         125|             0|         0|        65|     64197|         0|
|2018-05-27T13:56:47-04:00|         0|   131494277|         114|             0|         0|        58|     58542|         0|
+-------------------------+----------+------------+------------+--------------+----------+----------+----------+----------+

--- WiredTiger Summary ---
+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+
|                         | MaxBytes     | Currently    | Unmodified   | Tracked      | PagesRead    | PagesWritten |
| Date/Time               | Configured   | InCache      | PagesEvicted | DirtyBytes   | IntoCache    | FromCache    |
|-------------------------|--------------|--------------|--------------|--------------|--------------|--------------|
|2018-05-27T13:51:49-04:00|    1073741824|     371819216|      55696542|     223847654|          7536|        283333|
|2018-05-27T13:52:49-04:00|    1073741824|     932685796|      56085743|       1403512|        441614|         28656|
|2018-05-27T13:53:49-04:00|    1073741824|     894636138|      56718730|       1800916|        630904|            39|
|2018-05-27T13:54:49-04:00|    1073741824|     968276224|      58559554|        861193|       1845514|            55|
|2018-05-27T13:55:49-04:00|    1073741824|     953364800|      61252637|        929573|       2692357|            41|
|2018-05-27T13:56:47-04:00|    1073741824|     872418307|      63892617|       1162445|       2631892|            62|
+-------------------------+--------------+--------------+--------------+--------------+--------------+--------------+
2018/05/27 13:56:47 cleanup mongodb://localhost/?replicaSet=replset
2018/05/27 13:56:49 dropping collection _KEYHOLE_88800 keyhole
2018/05/27 13:56:49 dropping database _KEYHOLE_88800
```
