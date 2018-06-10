# Keyhole - MongoDB Performance Analytic
Keyhole is a performance measuring tool, written in GO (Golang), to collect stats from MongoDB instances and to measure performance of a MongoDB cluster.  It can also be configured as stats collecting agents and be expanded to a MongoDB instances monitoring tool.  Golang was chosen to eliminate the needs to install an interpreter (such as Java) or 3pp modules (such as Python or Node.js).  

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
- **Monitoring** mode to collect stats (see above) from `mongod` periodically.  Detail analytic results are displayed when the tool exists or can be viewed at a later time.
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

By default, it writes 2K size documents at 60 transactions per second from 10 different threads, a total of 600 TPS.  See sample outputs below.

```
Duration in minute(s): 5
...
Total TPS: 300 (tps) * 10 (conns) = 3000, duration: 5 (mins), bulk size: 512
...

2018-06-10T14:40:25-04:00 [replset] Memory - resident: 1067, virtual: 6100
2018-06-10T14:40:36-04:00 [replset] Storage: 460.6 -> 809.3, rate: 34.8 MB/sec
2018-06-10T14:40:46-04:00 [replset] Storage: 809.3 -> 1091.8, rate: 28.1 MB/sec
2018-06-10T14:40:56-04:00 [replset] Storage: 1091.8 -> 1375.6, rate: 28.2 MB/sec
2018-06-10T14:41:06-04:00 [replset] Storage: 1375.6 -> 1662.1, rate: 28.2 MB/sec
2018-06-10T14:41:16-04:00 [replset] Storage: 1662.1 -> 1869.0, rate: 20.6 MB/sec
```

### Load Test
Load test a cluster/replica.  A default cycle lasts five minutes with docs using in this [schema](LOADTEST.md).

- Populate data in first minute
- Perform CRUD operations during the second minutes
- Burst test until before the last minute
- Perform teardown ops in the last minute

```
keyhole --uri mongodb://localhost/?replicaSet=replset
```

It works on standalone, replica, and sharded cluster.  For a sharded cluster, *keyhole* collects stats from the primary node of all shards and display stats individually.  See [LOADTEST](LOADTEST.md) document for more details.

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
Populate a small amount of data to *\_KEYHOLE\_* database for [demo](SEED.md) and educational purposes such as CRUD, `$elemMatch`, `$lookup` (outer left join), indexes, and aggregation framework.

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
### Build
You need `go` installed and use `dep` to pull down dependencies.

```
cd $GOPATH/src
git clone https://github.com/simagix/keyhole.git
cd keyhole

dep ensure
go run keyhole.go --help

go build
./keyhole --help
```

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
