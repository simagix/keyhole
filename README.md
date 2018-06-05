# Keyhole - MongoDB Performance Analytic
Keyhole is a performance meansuring tool, written in GO (Golang), to collect stats from MongoDB instances and to meansure performance of a MongoDB cluster.  It can also be configured as stats collecting agents and be expanded to a MongoDB instances monitoring tool.  Golang was chosen to eliminate the needs to install an intepreter (such as Java) or 3pp modules (such as Python or Node.js).  

With Keyhole, experienced users should be able to spot performance issues and to determine whether upgrades are needed quickly from a few minutes of testing and analyzing the results.  Keyhole supports TLS/SSL connections.

Several features are available, and they are

- **Write Throughputs Test** measures the MongoDB performance by writing documents at a high rate to a MongoDB cluster.
- [**Load test**](LOADTEST.md) extends the *Write throughputs test* by issuing different ops against a MongoDB cluster.  Stats analytic is also provided
  - Memory: resident, virtual, and page faults
  - Executor and ops
  - Latency: read, write, and command
  - Metrics: index keys examined, collection scan, in-memory sort, and ops
  - WiredTiger analytic 
- Customized load test with a sample document.  Uses can load test using their own document format (see [LOADTEST.md](LOADTEST.md) for details).
- **Monitoring** mode to collcet stats (see above) from `mongod` periodically.  Detail analytic results are displayed when the tool exists or can be viewed at a later time.
- **Cluster Info** to display information of a cluster including stats to help determine physical memory size.
- [**Seed data**](SEED.md) for demo and educational purposes as a trainer.
- [Display average ops time](LOGINFO.md) and query patterns by parsing logs.

## Use Cases
Refer to [wiki](https://github.com/simagix/keyhole/wiki) for user's guide.

### Write Throughputs Test
Measure MongoDB write throughputs.

```
keyhole --uri mongodb://localhost/?replicaSet=replset --duration 1
```

By default, it writes 4K size documents at 600 transactions per second from 20 different threads.  See sample outputs below.

```
Total TPS: 600 (tps) * 20 (conns) = 12000, duration = 1 (mins)

2018-05-28T08:16:03-04:00 Memory - resident:     789, virtual:    5855
2018-05-28T08:16:13-04:00 Storage:  401.0 ->  730.4, rate   32.8 MB/sec
2018-05-28T08:16:23-04:00 Storage:  730.4 -> 1098.6, rate   36.5 MB/sec
2018-05-28T08:16:33-04:00 Storage: 1098.6 -> 1363.5, rate   26.4 MB/sec
2018-05-28T08:16:43-04:00 Storage: 1363.5 -> 1690.3, rate   31.4 MB/sec
```

### Load Test
Load test a cluster/replica.  A default cycle lasts five minutes with docs using in this [schema](LOADTEST.md).

- Populate data in first minute
- Perform CRUD operations during the second minutes
- Burst test until before the last minute
- Perform cleanup ops in the last minute

```
keyhole --uri mongodb://localhost/?replicaSet=replset
```

It works on standalone, replica, and sharded cluster.  However, for a sharded cluster, it only collects stats from one shard.  To collect stats from all shards, spin up different instances of `keyhole` and connect to each shard.  See [LOADTEST](LOADTEST.md) document for more details.

### Monitoring
Only collects data from `db.serverStatus()` command.  The [outputs](LOADTEST.md) share the same format from load test.

```
keyhole --uri mongodb://localhost/?replicaSet=replset --peek
```

Collected server status data is saved to a file and can be viewed later using the command below.

```
keyhole --view your_db_stats_file
```

### Cluster Info
Collect cluster information:

- Sharded cluster
- Replica set
- Standalone

```
keyhole --uri mongodb://localhost/?replicaSet=replset --info
```

**The command also displays total data and indexes sizes to help determine physical memory requirement from indexes and working set data size.**  Here is an example from a MongoDB Atlas cluster.

```
{
  "cluster": "replica",
  "host": "cluster0-shard-00-02-nhftn.mongodb.net:27017",
  "process": "mongod",
  "version": "3.6.4",
  "sharding": {},
  "repl": {
    "hosts": [
      "cluster0-shard-00-00-nhftn.mongodb.net:27017",
      "cluster0-shard-00-01-nhftn.mongodb.net:27017",
      "cluster0-shard-00-02-nhftn.mongodb.net:27017"
    ],
    "ismaster": false,
    "lastWrite": {
      "lastWriteDate": "2018-05-28T08:07:37-04:00",
      "majorityOpTime": {
        "t": 3,
        "ts": 6560602303152279000
      },
      "majorityWriteDate": "2018-05-28T08:07:37-04:00",
      "opTime": {
        "t": 3,
        "ts": 6560602303152279000
      }
    },
    "me": "cluster0-shard-00-02-nhftn.mongodb.net:27017",
    "primary": "cluster0-shard-00-00-nhftn.mongodb.net:27017",
    "rbid": 1,
    "secondary": true,
    "setName": "Cluster0-shard-0",
    "setVersion": 2
  },
  "TotalDBStats": {
    "statsDetails": [
      {
        "dataSize": 21191251,
        "db": "_KEYHOLE_",
        "indexSize": 2998272
      },
      {
        "dataSize": 0,
        "db": "admin",
        "indexSize": 0
      },
      {
        "dataSize": 0,
        "db": "local",
        "indexSize": 0
      }
    ],
    "totalDataSize": 21191251,
    "totalIndexSize": 2998272
  }
}
```

### Seed Data
Populate a small amount of data to *\_KEYHOLE\_* databse for [demo](SEED.md) and educational purposes such as CRUD, `$elemMatch`, `$lookup` (outer left join), indexes, and aggregation framework.

```
keyhole --uri mongodb://localhost/?replicaSet=replset --seed
```

Seeding data from a template is also supported, see [document](SEED.md) for details.

```
keyhole --uri mongodb://localhost/?replicaSet=replset --seed --file <json_file> [--drop]
```

### Ops Performance Analytic
Display ops average execution with query patterns using `--loginfo` flag.  See [LOGINFO.md](LOGINFO.md) for details.

```
keyhole --uri mongodb://localhost/?replicaSet=replset --loginfo ~/ws/demo/mongod.log
```

## Usages
### Download
Download the desired binary.  No other downloads (interpreter or modules) are necessary.  Please note that the builds of the master branch are changed often with new features.  For stable builds, use versioned branches.

#### MacOS
```
curl -L https://github.com/simagix/keyhole/raw/master/build/keyhole-osx-x64 > keyhole ; chmod +x keyhole
```
#### Linux
```
curl -L https://github.com/simagix/keyhole/raw/master/build/keyhole-linux-x64 > keyhole ; chmod +x keyhole
```
#### Windows
The download link is as below.

```
https://github.com/simagix/keyhole/raw/master/build/keyhole-win-x64.exe
```

### Usage
```
$ keyhole --help
```

### Atlas TLS/SSL Mode
An example connecting to Atlas

```
keyhole --uri mongodb://user:secret@cluster0-shard-00-01-nhftn.mongodb.net.:27017,cluster0-shard-00-02-nhftn.mongodb.net.:27017,cluster0-shard-00-00-nhftn.mongodb.net.:27017/test?replicaSet=Cluster0-shard-0\&authSource=admin --ssl --info
```
