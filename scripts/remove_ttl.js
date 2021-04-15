// Copyright 2021 Kuei-chun Chen. All rights reserved.
removeTTL();

// db.eventlog.createIndex( { "lastModifiedDate": 1 }, { expireAfterSeconds: 3600 } )
// db.serverlog.createIndex( { "expired_at": 1 }, { expireAfterSeconds: 0 } )

function removeTTL() {
  var removeCommands = [];
  var createCommands = [];
  var doc = db.adminCommand( { listDatabases: 1 } );
  var dbnames = doc["databases"];
  dbnames.forEach(function(database) {
    if(database["name"] == "admin" || database["name"] == "config" || database["name"] == "local") {
      // skip
    } else {
      var dbname = database["name"];
      var names = db.getSisterDB(dbname).getCollectionNames();
      names.sort();
      names.forEach(function(name) {
        var indexes = db.getSisterDB(dbname).getCollection(name).getIndexes();
        indexes.forEach(function(index) {
          if( index.hasOwnProperty("expireAfterSeconds") ) {
            var cmd = 'db.getSisterDB("'+dbname+'").'+name+'.dropIndex("'+index['name']+'");';
            removeCommands.push(cmd); 
            cmd = 'db.getSisterDB("'+dbname+'").'+name+'.createIndex(' +JSON.stringify(index['key'])+', { "expireAfterSeconds": '+index["expireAfterSeconds"]+' } );';
            createCommands.push(cmd);
          }
        });
      });
    }
  }); 

  print(); print("remove commands"); print();
  removeCommands.forEach(function(cmd) {
    print(cmd);
  });

  print(); print("create commands"); print();
  createCommands.forEach(function(cmd) {
    print(cmd);
  });
}
