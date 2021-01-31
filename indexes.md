# View Indexes Usages and Copy Indexes

*Keyhole* can print all indexes and their usages of collections of a database.  This is useful to evaluate redundant and/or unused indexes of a collection.  Another feature is to duplicate all indexes from a MongoDB cluster to another.

## List All Indexes and Usages

```
$ keyhole --index "mongodb+srv://user:secret@maobi-jgtm2.mongodb.net/"
maobi.numbers:
  { _id: 1 }
	host: maobi-shard-01-00-jgtm2.mongodb.net:27017, ops: 0, since: 2019-05-06 13:39:49.003 +0000 UTC
	host: maobi-shard-00-00-jgtm2.mongodb.net:27017, ops: 0, since: 2019-05-06 13:46:37.15 +0000 UTC
x { a: 1 }
	host: maobi-shard-01-00-jgtm2.mongodb.net:27017, ops: 0, since: 2019-05-06 13:40:29.677 +0000 UTC
	host: maobi-shard-00-00-jgtm2.mongodb.net:27017, ops: 0, since: 2019-05-06 13:46:37.172 +0000 UTC
x { a: 1, b: -1 }
	host: maobi-shard-01-00-jgtm2.mongodb.net:27017, ops: 0, since: 2019-05-06 13:40:26.085 +0000 UTC
	host: maobi-shard-00-00-jgtm2.mongodb.net:27017, ops: 0, since: 2019-05-06 13:46:37.168 +0000 UTC
x { a: -1, b: 1, c: -1 }
	host: maobi-shard-01-00-jgtm2.mongodb.net:27017, ops: 2, since: 2019-05-06 13:40:07.3 +0000 UTC
	host: maobi-shard-00-00-jgtm2.mongodb.net:27017, ops: 2, since: 2019-05-06 13:46:37.16 +0000 UTC
? { a: 1, b: 1, c: 1 }
	host: maobi-shard-01-00-jgtm2.mongodb.net:27017, ops: 0, since: 2019-05-06 13:40:13.979 +0000 UTC
	host: maobi-shard-00-00-jgtm2.mongodb.net:27017, ops: 0, since: 2019-05-06 13:46:37.164 +0000 UTC
x { a: 1, c: 1 }
	host: maobi-shard-01-00-jgtm2.mongodb.net:27017, ops: 0, since: 2019-05-06 13:40:43.344 +0000 UTC
	host: maobi-shard-00-00-jgtm2.mongodb.net:27017, ops: 0, since: 2019-05-06 13:46:37.177 +0000 UTC
* { b: 1 }
	host: maobi-shard-01-00-jgtm2.mongodb.net:27017, ops: 0, since: 2019-05-06 13:39:49.007 +0000 UTC
	host: maobi-shard-00-00-jgtm2.mongodb.net:27017, ops: 0, since: 2019-05-06 13:46:37.154 +0000 UTC
? { b: 1, c: 1 }
	host: maobi-shard-01-00-jgtm2.mongodb.net:27017, ops: 0, since: 2019-05-06 13:40:55.617 +0000 UTC
	host: maobi-shard-00-00-jgtm2.mongodb.net:27017, ops: 0, since: 2019-05-06 13:46:37.184 +0000 UTC
  { c: 1 }
	host: maobi-shard-01-00-jgtm2.mongodb.net:27017, ops: 1, since: 2019-05-06 13:40:47.214 +0000 UTC
	host: maobi-shard-00-00-jgtm2.mongodb.net:27017, ops: 1, since: 2019-05-06 13:46:37.18 +0000 UTC

Indexes info is written to maobi-shard-00-00-jgtm2.mongodb.net-index.bson.gz
```

Indexes to be reviewed, due to no usage, are in <span style="color:Blue">blue</span> with a lead '?'.  Duplicated indexes, which can be removed, are in <span style="color:Red">red</span> with a lead 'x'.  Indexes also used as shard keys are with leading '*'.


## Duplicate Indexes to Another MongoDB Cluster

The command with `--index` parameter outputs a file with a `-index.bson.gz` suffix.  Use the file and another cluster MongoDB connection string to duplicate indexes to the receiving cluster.  For example:

```
keyhole --createIndex maobi-shard-00-00-jgtm2.mongodb.net-index.bson.gz "mongodb+srv://user:secret@keyhole-jgtm2.mongodb.net/"
```