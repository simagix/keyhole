#! /bin/bash
# Copyright 2018 Kuei-chun Chen. All rights reserved.

echo ; echo "Spin up mongod"
mongod --version
mkdir -p data/db
rm -rf data/db/*
mongod --port 30097 --dbpath data/db --logpath data/mongod.log --fork --wiredTigerCacheSizeGB .5
mongo --port 30097 _KEYHOLE_88800 --eval "db.setProfilingLevel(0, {slowms: 20})"
export DATABASE_URL="mongodb://localhost:30097/"

# Test version
echo ; echo "==> Test version (--version)"
go run keyhole.go --version

# Test Schema
echo ; echo "==> Test printing default schema (--schema)"
go run keyhole.go --schema
echo ; echo "==> Test printing schema from a template (--schema --file <file>)"
go run keyhole.go --schema --file examples/template.json

# Test Info
echo ; echo "==> Test printing cluster info (--info <uri>)"
go run keyhole.go --info mongodb://localhost:30097/KEYHOLEDB

# Test seed
echo ; echo "==> Test seeding default docs (--seed <uri>)"
go run keyhole.go --seed mongodb://localhost:30097/KEYHOLEDB
echo ; echo "==> Test seeding default docs after dropping collection (--seed --drop <uri>)"
go run keyhole.go --seed --drop mongodb://localhost:30097/KEYHOLEDB
echo ; echo "==> Test seeding docs from a template (--file <file> --collection <collection> <uri>)"
go run keyhole.go --seed --file examples/template.json --collection template mongodb://localhost:30097/KEYHOLEDB
echo ; echo "==> Test seeding docs from a template after dropping collection (--file <file> --collection <collection> --drop <uri>)"
go run keyhole.go --seed --file examples/template.json --collection template --drop mongodb://localhost:30097/KEYHOLEDB

# Test load test
echo ; echo "==> Test load from a template (--file <file> <uri>)"
go run keyhole.go --file examples/template.json --duration 2 \
    --tps 300 --conn 10 --simonly mongodb://localhost:30097/KEYHOLEDB

go run keyhole.go --file examples/template.json --duration 3 \
    --tps 300 --conn 10 --tx examples/transactions.json mongodb://localhost:30097/KEYHOLEDB

# Test loginfo
echo ; echo "==> Test printing performance stats from a log file (--loginfo <file>)"
go run keyhole.go --loginfo data/mongod.log

echo ; echo "Shutdown mongod"
mongo --port 30097 --eval 'db.getSisterDB("admin").shutdownServer()'
rm -rf data/*
