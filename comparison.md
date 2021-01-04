# Clusters Comparison
The `-compare` feature compares two MongoDB clusters, either from connection strings or output files of `-allinfo`, and display results.  This feature provides a way of quick sanity check after migrating a MongoDB cluster to Atlas.

## Usage

```bash
keyhole [-nocolor] -compare <source_connection_string> <target_connection_string>
```

For example:

```bash
keyhole -compare 'mongodb://user:password@source.simagix.com?authSource=admin' 'mongodb+srv://user:password@target.mongodb.net/'
```

Or, compare two files:

```bash
keyhole -compare source-allinfo-stats.bson.gz target-allinfo-stats.bson.gz
```

## Example Outputs

The results are in red ink if the numbers between the source and the target are different.  If `-nocolor` flag is used, Keyhole mark it with a **≠** instead.

```bash
2021/01/02 15:39:44 === Comparison Results (source vs. target) ===
2021/01/02 15:39:44 Number of Databases:                   4               4
2021/01/02 15:39:44 Database keyhole
2021/01/02 15:39:44  ├─Number of Collections:              8               8
2021/01/02 15:39:44  ├─Number of Indexes:                 17              17 (all shards)
2021/01/02 15:39:44  ├─Number of Objects:              3,160           3,160
2021/01/02 15:39:44  ├─Total Data Size:                3.0MB           3.0MB
2021/01/02 15:39:44  ├─Average Data Size:                989             989
2021/01/02 15:39:44  └─Number of indexes:
2021/01/02 15:39:44    ├─keyhole.dealers:                  1               1
2021/01/02 15:39:44    ├─keyhole.employees:                1               1
2021/01/02 15:39:44    ├─keyhole.favorites:                1               1
2021/01/02 15:39:44    ├─keyhole.lookups:                  1               1
2021/01/02 15:39:44    ├─keyhole.models:                   1               1
2021/01/02 15:39:44    ├─keyhole.numbers:                  5               5
2021/01/02 15:39:44    ├─keyhole.robots:                   1               1
2021/01/02 15:39:44    ├─keyhole.vehicles:                 6               6
2021/01/02 15:39:44 Database maobi
2021/01/02 15:39:44  ├─Number of Collections:              8               8
2021/01/02 15:39:44  ├─Number of Indexes:                 18 ≠            17 (all shards)
2021/01/02 15:39:44  ├─Number of Objects:            401,149 ≠       401,147
2021/01/02 15:39:44  ├─Total Data Size:              584.5MB ≠       584.5MB
2021/01/02 15:39:44  ├─Average Data Size:              1.5KB ≠         1.5KB
2021/01/02 15:39:44  └─Number of indexes:
2021/01/02 15:39:44    ├─oplog.dealers:                    1               1
2021/01/02 15:39:44    ├─oplog.employees:                  1               1
2021/01/02 15:39:44    ├─oplog.favorites:                  1               1
2021/01/02 15:39:44    ├─oplog.lookups:                    1               1
2021/01/02 15:39:44    ├─oplog.models:                     1               1
2021/01/02 15:39:44    ├─oplog.numbers:                    5               5
2021/01/02 15:39:44    ├─oplog.robots:                     1               1
2021/01/02 15:39:44    ├─oplog.vehicles:                   6               6
2021/01/02 15:39:44 bson data written to ./out/hostname-compare.bson.gz
```