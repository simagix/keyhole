# Keyhole User's Guide
Keyhole is a MongoDB performance analytic tool and comes with many features. It can be overwhelmed when you first use the tool.  This document provides step by step instructions.

## Getting Started
Keyhole uses a MongoDB connection string to connect to a cluster.  For a standalone server, you typically have a connection string as follows.

```
mongodb://<usern>:<passwd>@<hostname>:<port>/?authSource=<db>
```

Once you have a connection string, connect to your cluster with `--uri` and `--info` flag.

```
keyhole --uri mongodb://user:password@localhost:27017/?authSource=admin --info
```

For a sharded cluster, connect to a `mongos` and the command is similar to connecting to a standalone server.  For a replica set, include all replica nodes and replica set name in the connection string.  Below is an example of connecting to a replica set.

```
keyhole --uri mongodb://user:password@localhost:27017,localhost:27018,localhost,27019/?replicaSet=replset\&authSource=admin --info
```

### TLS/SSL Support
If inflight encryption is required by your cluster, include `--ssl` and `--sslCAFile` in the command.  For example,

```
keyhole --uri mongodb://user:password@localhost:27017/?authSource=admin \
    --info --ssl --sslCAFile /etc/ssl/certs/ca.crt
```

In the following sections, I will cover use cases of how Keyhole is used for performance analytic.

## Memory
To determine the amount of physical memory should be installed, MongoDB recommends that for fastest processing, ensure that your [indexes fit entirely in RAM](https://docs.mongodb.com/manual/tutorial/ensure-indexes-fit-ram/) so that the system can avoid reading the index from disk.  The output of the *keyhole* command with `--info` flag has the storage information, and the total index size is the value of `StorageSize.totalIndexSize` in bytes.

## Seed Data
Seeding data is a very useful feature when you need need to populate data for demo or testing your applications.  By default, *keyhole* generates document from a predefined format for demo purpose.  But, *keyhole* can also generate data by mimicking a sample JSON document.

### Document Template
*Keyhole* can reads in a sample JSON document and populate data using randomized values.  This is especially useful for developers or QA to populate data before performing any tests.  For example, your have a document with following structure in file `template.json`.

```
{
	"_id": "a1b2c3d4e5f6f7e8d9c0b6a8",
	"email": "simagix@gmail.com",
	"hostIP": "192.168.1.1",
	"shortString": "Atlanta",
	"longString": "This is another string value. You can use any field names",
	"number": 123,
	"hex": "a1023b435c893d123e567f3487",
	"objectId": "a1b2c3d4e5f6f7e8d9c0b6a8",
	"lastUpdated": "2018-01-01T01:23:45Z",
	"array1": [123, 456, 789],
	"array2": [ "little", "cute", "girl" ],
	"array3": [
		{"city1": "New York", "city2": "Atlanta", "city3": "Miami"},
		{"city1": "Chicago", "city2": "Dallas", "city3": "Houston"} ],
	"subdocs": {
		"attribute1": {"email": "ken.chen@mongodb.com"}
	}
}
```

To view an example of document, execute the command below.

```
keyhole --schema --file template.json
```

A example of result is shown as following.

```
{
  "_id": ObjectId("de763b679dc88d11243b0bc2"),
  "array1": [
    8511,
    8162,
    5089
  ],
  "array2": [
    "Olivia",
    "Michael",
    "Olivia"
  ],
  "array3": [
    {
      "city1": "Ava",
      "city2": "Linda",
      "city3": "Ava"
    },
    {
      "city1": "Robert",
      "city2": "Mary",
      "city3": "Jennifer"
    }
  ],
  "email": "Richard.S.Jones@amazon.com",
  "hex": "dcd554c200aa94d6b11729b39b",
  "hostIP": "45.156.103.220",
  "lastUpdated": ISODate("1987-09-24T12:43:52-04:00"),
  "longString": "There's no place like home. All right, Mr. DeMille, I'm",
  "number": 563,
  "objectId": ObjectId("d739c08d26448db599b89278"),
  "shortString": "Richard",
  "subdocs": {
    "attribute1": {
      "email": "Liam.L.Smith@google.com"
    }
  }
}
```

Using a template, *keyhole* is intelligent enough to detect data types.  **More importantly, for the string type, *keyhole* can detect a few special formats, and they are email address, IP, date/time with `yyyy-MM-ddT\.*` format, and HEX, respectively and then replace their values with randomized values of their formats accordingly.**  Note that `ObjectID()` and `ISODate()` are not valid JSON syntax and you have to change them to a sub document using a 24 bytes HEX string and ISO date string format begins with "yyyy-MM-ddT", for example `2018-06-06T12:34:56Z`.

### Document Creation
*Keyhole* by default seeds 1,000 documents at a time to namespace `_KEYHOLE_.examples`

```
keyhole --uri mongodb://user:password@localhost:27017/?authSource=admin \
    --seed --file template.json
```

You can specify the destination database by including the target database in the connection string and the total number of documents to be created with the `--total` flags.  For example,

```
keyhole --uri mongodb://user:password@localhost:27017/EXAMPLEDB?authSource=admin \
    --seed --file template.json --total 5000
```

The above example writes 5,000 documents to `EXAMPLEDB.examples`.  In addition, *keyhole* inserts into the destination collection without dropping it first.  To drop the `_KEYHOLE_.examples` before any document insertion, include the `--drop` flag.

```
keyhole --uri mongodb://user:password@localhost:27017/?authSource=admin \
    --seed --file template.json --total 5000 --drop
```

## Monitoring
There are two ways to begin monitoring, and they are with `--peek` and `--monitor` options, respectively.

### Quick Snapshot `--peek`
*Keyhole* can collect server status data by issuing `db.serverStatus()` to a node or a cluster with the `--peek` flag.  if *keyhole* connects to a replica set, it will connect to the primary node.  Below is an example of connecting to a standalone instance.

```
keyhole --uri mongodb://user:password@localhost:27017/?authSource=admin --peek
```

Note that, the default database *keyhole* collects database stats from is `_KEYHOLE`.  You can include the targe database in the connection string as described in the *Seed Data* section.  By default, *keyhole* runs for a period of 5 minutes and intends to take a "snapshot" of the system performance.  To prolong the monitoring time, include the `--duration` flag.  For example, to monitor a cluster for 60 minutes, do

```
keyhole --uri mongodb://user:password@localhost:27017/?authSource=admin \
    --peek --duration 60
```

For a sharded cluster, *keyhole* connects to the primary node of all shards.  The reason of doing so, instead of remaining at `mongos`, is to collect more stats, such as wiredTiger stats data, and provide a partial view of the entire cluster.

### Stats Viewing
At the end of a *keyhole*, including monitoring and load test, it writes data to a file under system's default temporary directory.  Save the file and you can view the summaries again using *keyhole* with `--view` flag.  For example,

```
keyhole --view /tmp/keyhole_stats.2018-06-04T084021
```

### Monitor Mode `--monitor`
With `--monitor` flag, *keyhole* collects server status using `db.serverStats()` every **10 minutes** for, by default, **24 hours**.  At the end of each run, same as using `--peek`, *keyhole* prints a summary and stores server status data in a file.


```
keyhole --uri mongodb://user:password@localhost:27017/?authSource=admin --monitor
```

To change the default duration, include the `--duration` flag (in minutes).

```
keyhole --uri mongodb://user:password@localhost:27017/?authSource=admin \
    --monitor --duration 120
```

## Performance Analytic
*Keyhole* prints stats along the way during the run and summaries are displayed when *keyhole* exits or interrupted.  The discussion of what these metrics represent is outside the scope of this user's guide.  But experienced users should be able identify the performance bottlenecks from these summaries.

### Analytic Summary
Analytic summary includes

- Resident memory
- Virtual memory
- Page faults
- Number of ops from Command
- Number of ops from Delete
- Number of ops from Getmore
- Number of ops from Insert
- Number of ops from Query
- Number of ops from Update
- IOPS

*Page faults* occurs when `mongod` has to retrieved data from disks because data was not found in the memory.  High number of *Page faults* indicates the amount physical memory installed is not able to hold your working set.  Adding additional memory or scaling MongoDB horizontally is recommended.

### Global Lock Summary
Global lock includes

- Total time in milliseconds
- Active Client
- Current Queue

Growing number of *Current Queue* indicates MongoDB is not processing requests fast enough as they arrive. Scaling up your server can help, such as fast CPU and disk, or caling MongoDB horizontally by adding additional shards.

### Latencies Summary
Latencies summary includes

- reads latency in millisecond
- writes latency in millisecond
- commands latency in millisecond

### Metrics Summary
Metrics summary includes

- Scanned
- ScannedObj
- ScanAndOrder
- WriteConflicts
- Deleted
- Inserted
- Returned
- Updated

*Scanned* designates the total number of index items scanned during queries and query-plan evaluation.  *ScannedObj* indicates total number of documents scanned during queries and query-plan evaluation.  *ScanAndOrder* means the total number of queries that return sorted numbers that cannot perform the sort operation using an index.  Additional indexes can help if you have high numbers of *ScannedObj* or *ScanAndOrder*.

### WiredTiger Cache Summary
WireTiger cache summary includes

- MaxBytes Configured
- Currently InCache   
- Tracked DirtyBytes
- Modified PagesEvicted 
- Unmodified PagesEvicted     
- PagesRead IntoCache   
- PagesWritten FromCache

The value of *Currently InCache* should not be greater than the *MaxBytes Configured* value.  *Tracked DirtyBytes* indicates data in the wiredTiger cache that has been modified but not yet flushed to disk, and should be less than the value of *Currently in cache*. Expanding amounts of dirty data represents a bottleneck where data is written to disk not fast enough.  *Modified PagesEvicted* and *Unmodified PagesEvicted* reveal page eviction statistics.  *PagesRead IntoCache* and *PagesWritten FromCache* provide an overview of the I/O activity.

### WiredTiger Concurrent Transaction Summary
WireTiger cache summary includes

- Read ticket
- Write ticket

WiredTiger storage engine uses tickets to control the number of read/write operations simultaneously processed.  When the number of available read or write tickets available reaches zero, new read or write requests are queued until a new read or write ticket is freed up.

## Load Test
*Keyhole* can be used as a load testing tool on top of the collecting stats functions.  To load test a standalone server, simply execute it with the `--uri` flag.  Along the way, *keyhole* also collect server status as described in *Monitoring* section and prints summaries at the end of a run.

```
keyhole --uri mongodb://user:password@localhost:27017/?authSource=admin \
    --file template.json
```

The above command by default spawns 20 connections and generates 600 transactions per second (TPS) for 5 minutes.  You can change the number of connections by including the `--conn` flag, the TPS from `--tps` flag, and duration from `--duration` flag.

### Transactions Template
You can load test with pre-defined transactions and indexes with `--tx` flag.

```
keyhole --uri mongodb://user:password@localhost:27017/?authSource=admin \
    --file template.json --tx transactions.json
```

In the template file, you can define *indexes* and *transactions* as following example:

```
{
	"indexes": [{
		"email": 1, "hostIP": 1
	}],
	"transactions": [{
			"c": "insert"
		},
		{
			"c": "find",
			"q": {
				"email": "simagix@example.com"
			}
		},
		{
			"c": "update",
			"q": {
				"email": "simagix@example.com"
			},
			"o": {
				"$set": {
					"email": "simagix@example.com"
				}
			}
		}, {
			"c": "remove",
			"q": "{'email': 'simagix@example.com'}"
		}
	]
}
```

Both single field and compound indexes are supported.  In addition, supported commands are

- insert
- find
- findOne
- update
- updateAll
- remove
- removeAll

Load test a replica or a sharded cluster, the connecting behavior works the same way as in monitoring mode.  Note that under load test mode, *keyhole* ignores the target database in the connection string.  The reason is at the end of run *keyhole* will clean up data and thus it always run against the `_KEYHOLE_88800.keyhole` collection.

### Design
There are different stages in a *keyhole* load test run, and they are as following

#### Data Population
In the first minute of a load test run, *keyhole* insert documents into database.  The purposes of doing such are three folds.  First, we want to measure the throughputs of a MongoDB cluster in MB per second.  Secondly, by having the first minute activities, we can bring the `mongod` instances from "cold" to "hot".  Lastly, we should have enough data to work with at the end of the first minute.

#### Normal Ops
The minute following data population, it performances 4 IOPS (CRUD) per cycle for each transaction.  However, it only executes half of the TPS.  In other words, with 600 TPS, there are a total of 1,200 IOPS (4 * 600 / 2) per connection.

#### Thrashing
The period follows *normal ops* and before *teardown* are the thrashing period.  The system is in full throttle of 2,400 (4 * 600)IOPS per connection.  To determine if 600 TPS is too high for your cluster, include `--v` flag to turn on verbose mode.  "Simulate TPS overflows ..." messages will be displayed if you have a too high of a TPS.  In this case, you can lower the TPS or the number of connections.

#### Teardown
The last minute is the teardown period and it removes test data from database.

### Simulation Only Mode
*Keyhole* uses 600 TPS by default with a goal of measuring the maximum write throughputs of a cluster.  However, 600 TPS may be too high for some systems due to many factors, such as slower disk, jammed network, and document size.  If testing the bandwith of write throughputs is already done, with pre-populated data, you may want to run *keyhole* in a simulation only mode and lower the TPS.  For example,

```
keyhole --uri mongodb://user:password@localhost:27017/?authSource=admin \
    --file template.json --simonly --tps 400
```

## Log Analytic
*Keyhole* can identify slow operations and queries with collections scans by analyzing a log file.  Run *keyhole* command with a `--loginfo` flag and it will display average execution time by query patterns and sorted by execution in descending order.

```
keyhole --loginfo /var/log/mongodb/mongod.log
```
