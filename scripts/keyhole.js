// Copyright 2020 Kuei-chun Chen. All rights reserved.

var cluster = GetClustersSummary()
var data = JSON.stringify(cluster);
print(data)

function GetClustersSummary() {
  var cluster = { "cluster": "standalone", "config": {}, "databases": [] };
  GetServerInfo(cluster);
  var doc = db.adminCommand( { listDatabases: 1 } );
  dbnames = doc["databases"];
  dbnames.forEach(function(database) {
    if(database["name"] == "admin" || database["name"] == "config" || database["name"] == "local") {
      // skip
    } else {
      collections = [];
      var dbname = database["name"];
      names = db.getSisterDB(dbname).getCollectionNames()
      names.sort()
      names.forEach(function(name) {
        var firstDoc = db.getSisterDB(dbname).getCollection(name).findOne();
        var stats =	db.getSisterDB(dbname).runCommand( { "collStats": name });
        collections.push({"NS": dbname+"."+name, "collection": name, "document": firstDoc,
  				"indexes": GetIndexesFromCollection(dbname, name), "stats": stats});
      });
      var stats = db.getSisterDB(dbname).runCommand( { "dbStats": 1 });
      cluster["databases"].push({ "DB": dbname, "collections": collections, "stats": stats})
    }
  });
  return cluster;
}

function GetServerInfo(cluster) {
  doc = db.serverStatus();
  cluster["sharding"] = {};
  if (doc["process"] == "mongos") {
    cluster["cluster"] = "sharded"
    cluster["sharding"] = doc["sharding"];
    GetShards(cluster)
  } else if (doc["repl"] != null) {;
    cluster["cluster"] = "replica"
    cluster["config"]["oplog"] = GetOplogStats();
  } else {
    cluster["cluster"] = "standalone"
  }
  cluster["host"] = doc["host"];
  cluster["process"] = doc["process"];
  CollectServerInfo(cluster);
}

function GetOplogStats() {
  var oplog = {};
  var stats = db.getSisterDB("local").runCommand({ "collStats": "oplog.rs"});
	oplog["maxSize"] = Number(stats["maxSize"])
	oplog["size"] = stats["size"]
	oplog["count"] = stats["count"]
  var firstDoc = db.getSisterDB("local").getCollection("oplog.rs").find({}, {ts: 1}).sort({ "$natural": 1}).next();
  var lastDoc = db.getSisterDB("local").getCollection("oplog.rs").find({}, {ts: 1}).sort({ "$natural": -1}).next();
  oplog["durationInSeconds"] = DB.tsToSeconds(lastDoc.ts) - DB.tsToSeconds(firstDoc.ts);
  return oplog;
}

function GetShards(cluster) {
  var shards = db.getSisterDB("admin").runCommand({ "listShards": 1 });
  cluster["shardIDs"] = shards["shards"]
}

function CollectServerInfo(cluster) {
  cluster["config"]["hostInfo"] = db.getSisterDB("admin").runCommand({ "hostInfo": 1 });
  cluster["config"]["getCmdLineOpts"] = db.getSisterDB("admin").runCommand({ "getCmdLineOpts": 1 });
  cluster["config"]["buildInfo"] = db.getSisterDB("admin").runCommand({ "buildInfo": 1 });
  cluster["config"]["serverStatus"] = db.getSisterDB("admin").runCommand({ "serverStatus": 1 });
  cluster["config"]["usersInfo"] = db.getSisterDB("admin").runCommand({ "usersInfo": 1 });
  cluster["config"]["rolesInfo"] = db.getSisterDB("admin").runCommand({ "rolesInfo": 1 });
  if (cluster["cluster"] == "replica") {
    cluster["config"]["replSetGetStatus"] = db.getSisterDB("admin").runCommand({ "replSetGetStatus": 1 });
  }
}

function GetIndexesFromCollection(dbname, name) {
  var indexes = [];
  var istats = db.getSisterDB(dbname).getCollection(name).aggregate([ {"$indexStats": {}} ]);
  var docs = [];

  istats.forEach(function(doc) {
    docs.push(doc);
  });

  for (var n = 0; n < docs.length; n++) {
    var doc = docs[n];
    var index = { "totalOps": 0 };

    index["name"] = doc["name"];
    index["key"] = JSON.stringify(doc["key"]).replace(/"/g, "");
    var str = JSON.stringify(doc["key"]).replace(":-1", ":1").replace(/"/g, "");
    index["effectiveKey"] = str.substring(1, str.length-1);
    index["stats"] = [];
		for (var i = 0; i < docs.length; i++) {
			if (docs[i]["name"] == index["name"]) {
				index["totalOps"] += Number(docs[i]["accesses"]["ops"])
				index["stats"].push({ "host": docs[i]["host"], "accesses": {"since": docs[i]["accesses"]["since"], "ops": Number(docs[i]["accesses"]["ops"])}})
			}
		}
    indexes.push(index);
  }
  return indexes;
}
