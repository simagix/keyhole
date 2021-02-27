# Keyhole - Survey Your Mongo Land

![keyhole](/keyhole-logo-40x40.png) Keyhole is a performance analytics tool, written in GO (Golang), to collect stats and to analyze performance for MongoDB clusters.  Golang was chosen to eliminate the needs to install an interpreter or software modules.  Use [Maobi](https://www.simagix.com/2021/02/maobi-reports-generator-for-keyhole.html) to create HTML reports from data collected by Keyhole.

## Blogs

Peek at your MongoDB Clusters like a Pro with Keyhole:

- [Part 1](https://www.mongodb.com/blog/post/peek-at-your-mongodb-clusters-like-a-pro-with-keyhole-part-1)
- [Part 2](https://www.mongodb.com/blog/post/peek-at-your-mongodb-clusters-like-a-pro-with-keyhole-part-2)
- [Part 3](https://www.mongodb.com/blog/post/peek-your-clusters-like-pro-with-keyhole-part-3)

Updated post is available at [Survey Your Mongo Land](https://www.simagix.com/2021/02/survey-your-mongo-land.html).
## Use Cases

With Keyhole, experienced users should be able to spot performance issues and to determine whether upgrades are needed quickly from a few minutes of testing and analyzing the results.  Keyhole supports TLS/SSL connections.

Several features are available, and they are

- **Write Throughputs Test** measures the MongoDB performance by writing documents at a high rate to a MongoDB cluster.
- [**Load test**](simulation.md) extends the *Write throughputs test* by issuing different ops against a MongoDB cluster.  Stats analytic is also provided
  - Memory: resident, virtual, and page faults
  - Executor and ops
  - Latency: read, write, and command
  - Metrics: index keys examined, collection scan, in-memory sort, and ops
  - WiredTiger analytic
- Customized load test with a sample document.  Uses can load test using their own document format (see [simulation.md](simulation.md) for details).
- [Cluster Info and Sanity Check](https://www.simagix.com/2021/02/survey-your-mongo-land.html) to display information of a cluster including stats to help determine working set data size.
- [Display all indexes and their usages](https://www.simagix.com/2021/02/index-responsibly.html)
- [Duplicate all indexes](https://www.simagix.com/2021/02/index-responsibly.html) to another MongoDB cluster.
- [Seed data](https://github.com/simagix/keyhole/wiki/Seed-Data-using-a-Template) for demo and educational purposes as a trainer.
- [Display ops time](https://www.simagix.com/2021/02/feel-pulse-of-mongo.html) and query patterns by parsing logs.
- [Display indexes scores](https://github.com/simagix/keyhole/wiki/Indexes-Scores-and-Explain) of a query shape.
- [Monitor WiredTiger Cache](https://www.simagix.com/2021/02/peek-into-wiredtiger-cache.html) in near real time.
- [View FTDC Data and scores](https://github.com/simagix/keyhole/wiki/MongoDB-FTDC-and-Grafana-Integration) with a friendly interface.
- [MongoDB Atlas API](https://github.com/simagix/keyhole/wiki/Atlas-API) integration.

Refer to [wiki](https://github.com/simagix/keyhole/wiki) for user's guide.

## Build and Download

Build and download instructions are available at [Build and Download Keyhole](https://www.simagix.com/2021/02/build-and-download-keyhole_7.html).

## Usage

```bash
keyhole --help
```

## Unit Tests

```bash
./test.sh load
```

## License

[Apache-2.0 License](LICENSE)

## Disclaimer

This software is not supported by MongoDB, Inc. under any of their commercial support subscriptions or otherwise. Any usage of keyhole is at your own risk. Bug reports, feature requests and questions can be posted in the Issues section on GitHub.
