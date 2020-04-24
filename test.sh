#! /bin/bash
# Copyright 2018 Kuei-chun Chen. All rights reserved.

shutdownServer() {
    echo "Shutdown mongod"
    mongo --quiet --port 30097 --eval 'db.getSisterDB("admin").shutdownServer()' > /dev/null 2>&1
    rm -rf data/db data/mongod.log.*
    exit
}

validate() {
  if [ $? != 0 ]; then
      echo $1
      shutdownServer
  fi
}

echo ; echo "Spin up mongod"
mongod --version
mkdir -p data/db
rm -rf data/db/*
mongod --port 30097 --dbpath data/db --logpath data/mongod.log --fork --wiredTigerCacheSizeGB 1  --replSet replset
validate "failed to start mongod"
mongo --quiet mongodb://localhost:30097/admin --eval 'rs.initiate()'
sleep 2
mongo --port 30097 _KEYHOLE_88800 --eval "db.setProfilingLevel(0, {slowms: 5})"
validate "failed to set profiling level"

export DATABASE_URI="mongodb://localhost:30097/keyhole?replicaSet=replset"

# Test version
echo ; echo "==> Test version (--version)"
go run keyhole.go --version
validate ""

# Test Info
echo ; echo "==> Test printing cluster info (--info <uri>)"
go run keyhole.go --info $DATABASE_URI
if [ $? != 0 ]; then
    exit
fi

# Test seed
echo ; echo "==> Test seeding default docs (--seed <uri>)"
go run keyhole.go --seed $DATABASE_URI
validate ""

echo ; echo "==> Test seeding default docs after dropping collection (--seed --drop <uri>)"
go run keyhole.go --seed --drop $DATABASE_URI
validate ""

mongo $DATABASE_URI --eval 'db.cars.createIndex({color: 1})'
mongo $DATABASE_URI --eval 'db.cars.createIndex({color: 1, style: 1})'

echo ; echo "==> Test seeding docs from a template (--file <file> --collection <collection> <uri>)"
go run keyhole.go --seed --file examples/template.json --collection template $DATABASE_URI
validate ""

echo ; echo "==> Test seeding docs from a template after dropping collection (--file <file> --collection <collection> --drop <uri>)"
go run keyhole.go --seed --file examples/template.json --collection template --drop $DATABASE_URI
validate ""

# Test Index
echo ; echo "==> Test printing cluster indexes (--index <uri>)"
go run keyhole.go --index $DATABASE_URI
validate ""

# Test Schema
echo ; echo "==> Test printing schema from a template (--schema --collection <collection> <uri>)"
go run keyhole.go --schema --collection cars $DATABASE_URI
validate ""

# Test Cardinality
echo ; echo "==> Test printing number of distinct fileds values (--cardinality)"
go run keyhole.go --cardinality favorites $DATABASE_URI
validate ""

# Test Cardinality
echo ; echo "==> Test printing number of distinct fileds values (--explain)"
go run keyhole.go --explain mdb/testdata/cars.log $DATABASE_URI
validate ""

# Test Cardinality
echo ; echo "==> Test printing number of distinct fileds values (--explain)"
go run keyhole.go --explain mdb/testdata/cars.json $DATABASE_URI
validate ""

if [ "$1" != "" ]; then
    # Test load test
    echo ; echo "==> Test load from a template (--file <file> <uri>)"
    go run keyhole.go --file examples/template.json --duration 2 \
        --tps 300 --conn 10 --simonly $DATABASE_URI
    validate ""

    go run keyhole.go --file examples/template.json --duration 3 \
        --tps 300 --conn 10 --tx examples/transactions.json $DATABASE_URI
    validate ""

    # Test loginfo
    echo ; echo "==> Test printing performance stats from a log file (--loginfo <file>)"
    go run keyhole.go --loginfo data/mongod.log
    rm -f *-mongod.log.gz
fi

# Test info Atlas
echo ; echo "==> Test printing clusters summary (--info <atlas_uri>)"
go run keyhole.go --info "atlas://${ATLAS_AUTH}"

# Test loginfo Atlas
# echo ; echo "==> Test printing performance stats from a log file (--loginfo <atlas_uri>)"
# go run keyhole.go --seed ${ATLAS_URL}
# go run keyhole.go --loginfo "atlas://${ATLAS_AUTH}@${ATLAS_GROUP}/keyhole"
rm -f mongodb.log.*

shutdownServer
