# Logs Analytics

*Keyhole* can identify slow operations and queries with collections scans by analyzing a log file, either plain or gzip compressed.  Run *keyhole* command with a `--loginfo` flag and it will display average execution time by query patterns and sorted by execution in descending order.  

## Usage
```
keyhole --loginfo [--collscan] [-v] [--regex {regular expression}] [-redact] logfile[.gz] [more log files]
```

## Examples
```
keyhole --loginfo /var/log/mongodb/mongod.log.2018-06-07T11-08-32.gz
```

With `--collscan` flag, *keyhole* only prints those with *COLLSCAN*.

```
keyhole --loginfo --collscan /var/log/mongodb/mongod.log.2018-06-07T11-08-32.gz
```

With `-v` flag, *keyhole* prints indexes associated with each query.

```
keyhole --loginfo -v /var/log/mongodb/mongod.log.2018-06-07T11-08-32.gz
```
## Customer Regular Expression
For those forward logs to syslogd, you can still parse the logs by including a regular expression.  Below is an example:

```
keyhole --loginfo --regex '^.*\[(\S+)\]:\s+\[\w+\] (\w+) (\S+) \S+: (.*) (\d+)ms$' syslog.4.log.gz
```

Note that your case may use a different regular expression.

## Sample outputs

```
+----------+--------+------+--------+------+---------------------------------+--------------------------------------------------------------+
| Command  |COLLSCAN|avg ms| max ms | Count| Namespace                       | Query Pattern                                                |
|----------+--------+------+--------+------+---------------------------------+--------------------------------------------------------------|
|insert                 862     1063      9 keyhole.favorites                 N/A                                                           |
|insert                 157      157      1 keyhole.robots                    N/A                                                           |
|insert                 115      115      1 keyhole.numbers                   N/A                                                           |
|update                 115      115      1 keyhole.employees                 {"_id":1}                                                     |
|...index:  IDHACK                                                                                                                          |
|insert                 107      107      1 keyhole.models                    N/A                                                           |
|insert                  30     1883   6873 keyhole.__examples                N/A                                                           |
|find                     3        3      3 keyhole.cars                      {"color":/.../}                                               |
|...index:  { color: 1 }                                                                                                                    |
|find                     1        1      3 keyhole.cars                      {"color":/.../i}                                              |
|...index:  { color: 1 }                                                                                                                    |
|find       COLLSCAN      1       47     50 keyhole.dealers                   {}                                                            |
+----------+--------+------+--------+------+---------------------------------+--------------------------------------------------------------+

top 10 of 35 lines displayed; see HTML report for details.
bson log info written to ./out/mongod.json-log.bson.gz
TSV log info written to ./out/mongod.json.tsv
```

## Report Generating

Drag the output file, `mongod-log.bson.gz`, to [Maobi](https://hub.docker.com/repository/docker/simagix/maobi) to create an HTML report.