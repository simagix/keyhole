# Clusters Comparison
The `-compare` feature compares two MongoDB clusters, either from connection strings or output files of `-allinfo`, and display results.  This feature provides a way of quick sanity check after migrating to a MongoDB cluster to Atlas.

## Usage

```bash
keyhole -compare <source_connection_string> <target_connection_string>
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

```bash
2020/12/31 22:15:13 
=== Comparison Results (source vs. target) ===
2020/12/31 22:15:13 Number of Databases:                   4               5
2020/12/31 22:15:13 Database keyhole
2020/12/31 22:15:13  ├─Number of Collections:              8               8
2020/12/31 22:15:13  ├─Number of Indexes:                 17              17 (all shards)
2020/12/31 22:15:13  ├─Number of Objects:              3,160           3,160
2020/12/31 22:15:13  ├─Total Data Size:                3.0MB           3.0MB
2020/12/31 22:15:13  ├─Average Data Size:                989             989
2020/12/31 22:15:13  └─Number of indexes
2020/12/31 22:15:13    ├─keyhole.dealers:                  1               1
2020/12/31 22:15:13    ├─keyhole.employees:                1               1
2020/12/31 22:15:13    ├─keyhole.favorites:                1               1
2020/12/31 22:15:13    ├─keyhole.lookups:                  1               1
2020/12/31 22:15:13    ├─keyhole.models:                   1               1
2020/12/31 22:15:13    ├─keyhole.numbers:                  5               5
2020/12/31 22:15:13    ├─keyhole.robots:                   1               1
2020/12/31 22:15:13    ├─keyhole.vehicles:                 6               6
2020/12/31 22:15:13 Database maobi
2020/12/31 22:15:13  ├─Number of Collections:              8               8
2020/12/31 22:15:13  ├─Number of Indexes:                 23              17 (all shards)
2020/12/31 22:15:13  ├─Number of Objects:             21,139          21,139
2020/12/31 22:15:13  ├─Total Data Size:               29.3MB          29.3MB
2020/12/31 22:15:13  ├─Average Data Size:              1.4KB           1.4KB
2020/12/31 22:15:13  └─Number of indexes
2020/12/31 22:15:13    ├─maobi.dealers:                    1               1
2020/12/31 22:15:13    ├─maobi.employees:                  1               1
2020/12/31 22:15:13    ├─maobi.favorites:                  1               1
2020/12/31 22:15:13    ├─maobi.lookups:                    1               1
2020/12/31 22:15:13    ├─maobi.models:                     1               1
2020/12/31 22:15:13    ├─maobi.numbers:                    5               5
2020/12/31 22:15:13    ├─maobi.robots:                     1               1
2020/12/31 22:15:13    ├─maobi.vehicles:                   6               6
2020/12/31 22:15:13 bson data written to ./out/hostname-compare.bson.gz
```