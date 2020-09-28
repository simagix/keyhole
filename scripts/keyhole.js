// Copyright 2020 Kuei-chun Chen. All rights reserved.
var cluster = GetClustersSummary()
var data = JSON.stringify(cluster);
print(data);

function GetClustersSummary() {
  var cluster = { "cluster": "standalone", "databases": [], 
    "keyhole": { "version": "v2.5.0-js", "collected": new Date(), "logs": [] } };
  cluster["keyhole"]["logs"].push(new Date().toISOString()+" keyhole.js began");
  var doc = db.serverStatus();
  cluster["host"] = doc["host"];
  cluster["process"] = doc["process"];
  cluster["hostInfo"] = db.getSisterDB("admin").runCommand({ "hostInfo": 1 });
  cluster["getCmdLineOpts"] = db.getSisterDB("admin").runCommand({ "getCmdLineOpts": 1 });
  cluster["buildInfo"] = db.getSisterDB("admin").runCommand({ "buildInfo": 1 });
  cluster["serverStatus"] = db.getSisterDB("admin").runCommand({ "serverStatus": 1 });
  if (cluster["cluster"] == "replica") {
    cluster["replSetGetStatus"] = db.getSisterDB("admin").runCommand({ "replSetGetStatus": 1 });
  }
  if (doc["process"] == "mongos") {
    cluster["cluster"] = "sharded";
    var shards = db.getSisterDB("admin").runCommand({ "listShards": 1 });
    cluster["shards"] = shards["shards"];
  } else if (doc["repl"] != null) {
    cluster["cluster"] = "replica";
    var node = { "host": doc["host"], "oplog": GetOplogStats() };
    var shard = { "_id": doc["repl"]["setName"], "servers": [] };
    node["buildInfo"] = cluster["buildInfo"];
    node["hostInfo"] = cluster["hostInfo"];
    node["serverStatus"] = cluster["serverStatus"];
    shard.servers.push(node);
    cluster["shards"] = [];
    cluster["shards"].push(shard); 
  } else {
    cluster["cluster"] = "standalone";
  }
  cluster["keyhole"]["logs"].push(new Date().toISOString()+" began GetDatabases()");
  cluster["databases"] = GetDatabases();
  cluster["keyhole"]["logs"].push(new Date().toISOString()+" ended GetDatabases()");
  cluster["keyhole"]["logs"].push(new Date().toISOString()+" keyhole.js exiting");
  return cluster;
}

function GetOplogStats() {
  var oplog = {};
  var stats = db.getSisterDB("local").runCommand({ "collStats": "oplog.rs"});
	oplog["maxSize"] = Number(stats["maxSize"]);
	oplog["size"] = stats["size"];
	oplog["count"] = stats["count"];
  var firstDoc = db.getSisterDB("local").getCollection("oplog.rs").find({}, {ts: 1}).sort({ "$natural": 1}).next();
  var lastDoc = db.getSisterDB("local").getCollection("oplog.rs").find({}, {ts: 1}).sort({ "$natural": -1}).next();
  oplog["durationInSeconds"] = DB.tsToSeconds(lastDoc.ts) - DB.tsToSeconds(firstDoc.ts);
  return oplog;
}

function GetDatabases() {
  var databases = [];
  var doc = db.adminCommand( { listDatabases: 1 } );
  var dbnames = doc["databases"];
  dbnames.forEach(function(database) {
    if(database["name"] == "admin" || database["name"] == "config" || database["name"] == "local") {
      // skip
    } else {
      collections = [];
      var dbname = database["name"];
      names = db.getSisterDB(dbname).getCollectionNames();
      names.sort();
      names.forEach(function(name) {
        var firstDoc = db.getSisterDB(dbname).getCollection(name).findOne();
        var stats =	db.getSisterDB(dbname).runCommand( { "collStats": name });
        collections.push({"namespace": dbname+"."+name, "name": name, "document": firstDoc,
          "indexes": GetIndexesFromCollection(dbname, name), "stats": stats});
      });
      var stats = db.getSisterDB(dbname).runCommand( { "dbStats": 1 });
      databases.push({ "name": dbname, "collections": collections, "stats": stats});
    }
  });
  return databases;
}

function GetIndexesFromCollection(dbname, name) {
  var indexes = [];
  var istats = db.getSisterDB(dbname).getCollection(name).aggregate([ {"$indexStats": {}} ]);
  var docs = [];
  istats.forEach(function(doc) {
    docs.push(doc);
  });

  for(var i = 0; i < docs.length; i++) {
    var doc = docs[i]; 
    var exists = false;

    for(var j = 0; j < indexes.length; j++) {
      var index = indexes[j];
      if (index["name"] == doc["name"] ) {  // found
        var access = doc["accesses"];
        indexes[j]["totalOps"] += Number(access["ops"]) 
        indexes[j]["usage"].push({ "host": doc["host"], "accesses": {"since": access["since"], "ops": Number(access["ops"])}});
        exists = true;
        break;
      }
    }

    if (exists == false) { // not found, push
      var access = doc["accesses"];
      doc = doc["spec"];
      doc["keystring"] = JSON.stringify(doc["key"]).replace(/"/g, "");
      var str = JSON.stringify(doc["keystring"]).replace(":-1", ":1").replace(/"/g, "");
      doc["effectiveKey"] = str.substring(1, str.length-1);
      doc["totalOps"] = Number(access["ops"]);
      doc["usage"] = [];
      doc["usage"].push({ "host": doc["host"], "accesses": {"since": access["since"], "ops": Number(access["ops"])}});
      indexes.push(doc);
    }
  }
  return indexes;
}
