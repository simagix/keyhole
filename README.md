# Keyhole - MongoDB Performance Analytic
Keyhole is a performance analytics tool, written in GO (Golang), to collect stats from MongoDB instances and to measure performance of a MongoDB cluster.  Moreover, keyhole can read MongoDB full-time diagnostic data (FTDC) data and is [integrated with Grafana's Simple JSON plugin](https://github.com/simagix/keyhole/wiki/MongoDB-FTDC-and-Grafana-Integration) seamlessly.  Golang was chosen to eliminate the needs to install an interpreter or 3pp modules.  [Download executable](https://github.com/simagix/keyhole#download).

With Keyhole, experienced users should be able to spot performance issues and to determine whether upgrades are needed quickly from a few minutes of testing and analyzing the results.  Keyhole supports TLS/SSL connections.

Several features are available, and they are

- **Write Throughputs Test** measures the MongoDB performance by writing documents at a high rate to a MongoDB cluster.
- [**Load test**](docs/LOADTEST.md) extends the *Write throughputs test* by issuing different ops against a MongoDB cluster.  Stats analytic is also provided
  - Memory: resident, virtual, and page faults
  - Executor and ops
  - Latency: read, write, and command
  - Metrics: index keys examined, collection scan, in-memory sort, and ops
  - WiredTiger analytic
- Customized load test with a sample document.  Uses can load test using their own document format (see [LOADTEST.md](docs/LOADTEST.md) for details).
- **Cluster Info** to display information of a cluster including stats to help determine physical memory size.
- [Display all indexes and their usages](https://github.com/simagix/keyhole/wiki/List-All-Indexes-with-Usages)
- [**Seed data**](https://github.com/simagix/keyhole/wiki/Seed-Data-using-a-Template) for demo and educational purposes as a trainer.
- [Display average ops time](https://github.com/simagix/keyhole/wiki/Mongo-Logs-Analytics) and query patterns by parsing logs.

## Use Cases
Refer to [wiki](https://github.com/simagix/keyhole/wiki) for user's guide.

### Write Throughputs Test
Measure MongoDB write throughputs.

```
keyhole --duration 1 mongodb://localhost/?replicaSet=replset
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
Load test a cluster/replica.  A default cycle lasts five minutes with docs using in this [schema](docs/LOADTEST.md).

- Populate data in first minute
- Perform CRUD operations during the second minutes
- Burst test until before the last minute
- Perform teardown ops in the last minute

```
keyhole mongodb://localhost/?replicaSet=replset
```

It works on standalone, replica, and sharded cluster.  For a sharded cluster, *keyhole* collects stats from the primary node of all shards and display stats individually.  See [LOADTEST](docs/LOADTEST.md) document for more details.

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
Download the desired binary.  No other downloads (interpreter or modules) are necessary.

#### Linux
```
curl -L https://github.com/simagix/keyhole/raw/master/build/keyhole-linux-x64 > keyhole ; chmod +x keyhole
```

#### MacOS
Download [keyhole](https://github.com/simagix/keyhole/raw/master/build/keyhole-win-x64.exe) for macOS, or,

```
curl -L https://github.com/simagix/keyhole/raw/master/build/keyhole-osx-x64 > keyhole ; chmod +x keyhole
```

#### Windows
Download [Windows  executable](https://github.com/simagix/keyhole/raw/master/build/keyhole-win-x64.exe).

### Usage
```
$ keyhole --help
```

### Atlas TLS/SSL Mode
An example connecting to Atlas

```
keyhole --info "mongodb+srv://user:secret@cluster0-v7due.gcp.mongodb.net/test"
```

or

```
keyhole --info "mongodb://user:secret@cluster0-shard-00-00-v7due.gcp.mongodb.net:27017,cluster0-shard-00-01-v7due.gcp.mongodb.net:27017,cluster0-shard-00-02-v7due.gcp.mongodb.net:27017/test?replicaSet=Cluster0-shard-0&authSource=admin&ssl=true"
```

### TLS/SSL Mode
```
keyhole --info --sslCAFile /etc/ssl/certs/ca.pem --sslPEMKeyFile /etc/ssl/certs/client.pem "mongodb://user:password@localhost/keyhole?authSource=admin&ssl=true"
```
