# Keyhole
Peek into `mongod` for

- Write Throughputs Test
- [Load test](LOADTEST.md)
- Monitoring
- Cluster Info
- [Seed data](SEED.md)

## Use Cases
### Write Throughputs Test
Test MongoDB write throughput.  

```
build/keyhole-linux-x64 -uri=mongodb://localhost/?replicaSet=replset -duration=1
```

### Load Test
Load test a cluster/replica.  A default cycle last six minutes.

- Populate data in first minute
- Perform CRUD operations during the second and third minutes
- Burst test during the fourth and fifth minutes
- Perform CRUD ops in the last minute

```
build/keyhole-linux-x64 -uri=mongodb://localhost/?replicaSet=replset
```

It works on standalone, replica, and sharded cluster.  However, for a sharded cluster, it only collects stats from one shard.  To collect stats from all shards, spin up different instances of `keyhole` and connect to each shard.

### Monitoring
Only collects data from `db.serverStatus()`

```
build/keyhole-linux-x64 -uri=mongodb://localhost/?replicaSet=replset -peek
```

### Cluster Info
Collect cluster information:

- Sharded cluster
- Replica set
- Standalone

```
build/keyhole-linux-x64 -uri=mongodb://localhost/?replicaSet=replset -info
```

### Seed Data
Populate a small amount of data for demo.

```
build/keyhole-linux-x64 -uri=mongodb://localhost/?replicaSet=replset -seed
```

## Usage
```
$ build/keyhole-linux-x64 -h
  -conn int
    	nuumber of connections (default 20)
  -duration int
    	load test duration in minutes (default 6)
  -info
    	get cluster info
  -peek
    	only collect data
  -seed
    	seed a database for demo
  -ssl
    	use TLS/SSL
  -sslCAFile string
    	CA file
  -tps int
    	number of trasaction per second per connection (default 600)
  -uri string
    	Mongo
    	DB URI
  -v	verbose
  -view string
    	server status file
```

## Download
### MacOS
```
curl -L https://github.com/simagix/keyhole/blob/master/build/keyhole-osx-x64?raw=true > keyhole ; chmod +x keyhole
```
### Linux
```
curl -L https://github.com/simagix/keyhole/blob/master/build/keyhole-linux-x64?raw=true > keyhole ; chmod +x keyhole
```
### Windows
The download link is as below.

```
https://github.com/simagix/keyhole/blob/master/build/keyhole-win-x64.exe?raw=true
```

## Atlas TLS/SSL Mode
An example connecting to Atlas

```
build/keyhole-osx-x64 -uri=mongodb://user:secret@cluster0-shard-00-01-nhftn.mongodb.net.:27017,cluster0-shard-00-02-nhftn.mongodb.net.:27017,cluster0-shard-00-00-nhftn.mongodb.net.:27017/test?replicaSet=Cluster0-shard-0\&authSource=admin -ssl -sslCAFile=ssl/ca.crt -info
```
