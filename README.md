# Keyhole - Survery Your Mongo Land

Keyhole is a performance analytics tool, written in GO (Golang), to collect stats from MongoDB instances and to analyze performance of a MongoDB cluster.  Golang was chosen to eliminate the needs to install an interpreter or software modules.  To generate HTML reports use [Maobi](https://hub.docker.com/repository/docker/simagix/maobi), a Keyhole reports generator.

## Blogs

Peek at your MongoDB Clusters like a Pro with Keyhole

- [Part 1](https://www.mongodb.com/blog/post/peek-at-your-mongodb-clusters-like-a-pro-with-keyhole-part-1)
- [Part 2](https://www.mongodb.com/blog/post/peek-at-your-mongodb-clusters-like-a-pro-with-keyhole-part-2)
- [Part 3](https://www.mongodb.com/blog/post/peek-your-clusters-like-pro-with-keyhole-part-3)

## Use Cases

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
- [Cluster Info and Sanity Check](https://github.com/simagix/keyhole/wiki/MongoDB-Cluster-Info) to display information of a cluster including stats to help determine working set data size.
- [Display all indexes and their usages](https://github.com/simagix/keyhole/wiki/View-Indexes-Usages-and-Copy-Indexes)
- [Duplicate all indexes](https://github.com/simagix/keyhole/wiki/View-Indexes-Usages-and-Copy-Indexes) to another MongoDB cluster.
- [**Seed data**](https://github.com/simagix/keyhole/wiki/Seed-Data-using-a-Template) for demo and educational purposes as a trainer.
- [Display average ops time](https://github.com/simagix/keyhole/wiki/Logs-Analytics) and query patterns by parsing logs.
- [Display indexes scores](https://github.com/simagix/keyhole/wiki/Indexes-Scores-and-Explain) of a query shape.
- [Monitor WiredTiger Cache](https://github.com/simagix/keyhole/wiki/WiredTiger-Cache-Usage) in near real time.
- [View FTDC Data and scores](https://github.com/simagix/keyhole/wiki/MongoDB-FTDC-and-Grafana-Integration) with a friendly interface.
- [MongoDB Atlas API](https://github.com/simagix/keyhole/wiki/Atlas-API) integration.

Refer to [wiki](https://github.com/simagix/keyhole/wiki) for user's guide.

## Build

You need `go` installed and use `dep` to pull down dependencies.

```bash
./build.sh
```

## Usage

```bsdh
keyhole --help
```

## Unit Tests

```bash
./test.sh load
```

## Atlas TLS/SSL Mode

An example connecting to Atlas

```bash
keyhole --info "mongodb+srv://user:secret@cluster0-v7due.gcp.mongodb.net/test"
```

## TLS/SSL Mode

```bash
keyhole --info --sslCAFile /etc/ssl/certs/ca.pem --sslPEMKeyFile /etc/ssl/certs/client.pem "mongodb://user:password@localhost/keyhole?authSource=admin&ssl=true"
```
