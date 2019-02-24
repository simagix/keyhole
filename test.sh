#! /bin/bash
# Copyright 2018 Kuei-chun Chen. All rights reserved.

shutdownServer() {
    echo "Shutdown mongod"
    mongo --quiet --port 30097 --eval 'db.getSisterDB("admin").shutdownServer()' > /dev/null 2>&1
    rm -rf data/db data/mongod.log.*.gz
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
mongo --port 30097 _KEYHOLE_88800 --eval "db.setProfilingLevel(0, {slowms: 10})"
validate "failed to set profiling level"

export DATABASE_URL="mongodb://localhost:30097/"

# Test version
echo ; echo "==> Test version (--version)"
go run keyhole.go --version
validate ""

# Test Info
echo ; echo "==> Test printing cluster info (--info <uri>)"
go run keyhole.go --info mongodb://localhost:30097/KEYHOLEDB?replicaSet=replset
if [ $? != 0 ]; then
    exit
fi

# Test seed
echo ; echo "==> Test seeding default docs (--seed <uri>)"
go run keyhole.go --seed mongodb://localhost:30097/KEYHOLEDB?replicaSet=replset
validate ""

echo ; echo "==> Test seeding default docs after dropping collection (--seed --drop <uri>)"
go run keyhole.go --seed --drop mongodb://localhost:30097/KEYHOLEDB?replicaSet=replset
validate ""

echo ; echo "==> Test seeding docs from a template (--file <file> --collection <collection> <uri>)"
go run keyhole.go --seed --file examples/template.json --collection template mongodb://localhost:30097/KEYHOLEDB?replicaSet=replset
validate ""

echo ; echo "==> Test seeding docs from a template after dropping collection (--file <file> --collection <collection> --drop <uri>)"
go run keyhole.go --seed --file examples/template.json --collection template --drop mongodb://localhost:30097/KEYHOLEDB?replicaSet=replset
validate ""

# Test Index
echo ; echo "==> Test printing cluster indexes (--index <uri>)"
go run keyhole.go --index mongodb://localhost:30097/KEYHOLEDB?replicaSet=replset
validate ""

# Test Schema
echo ; echo "==> Test printing default schema (--schema)"
go run keyhole.go --schema
validate ""

echo ; echo "==> Test printing schema from a template (--schema --file <file>)"
go run keyhole.go --schema --file examples/template.json
validate ""

echo ; echo "==> Test printing schema from a template (--schema --collection <collection> <uri>)"
go run keyhole.go --schema --collection favorites mongodb://localhost:30097/KEYHOLEDB?replicaSet=replset
validate ""

# Test Cardinality
echo ; echo "==> Test printing number of distinct fileds values (--card)"
go run keyhole.go --card --collection favorites mongodb://localhost:30097/KEYHOLEDB?replicaSet=replset
validate ""

# Test load test
echo ; echo "==> Test load from a template (--file <file> <uri>)"
go run keyhole.go --file examples/template.json --duration 2 \
    --tps 300 --conn 10 --simonly mongodb://localhost:30097/KEYHOLEDB?replicaSet=replset
validate ""

go run keyhole.go --file examples/template.json --duration 3 \
    --tps 300 --conn 10 --tx examples/transactions.json mongodb://localhost:30097/KEYHOLEDB?replicaSet=replset
validate ""

# Test loginfo
echo ; echo "==> Test printing performance stats from a log file (--loginfo <file>)"
go run keyhole.go --loginfo data/mongod.log
shutdownServer
