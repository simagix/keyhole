#! /bin/bash
# Copyright 2018 Kuei-chun Chen. All rights reserved.

echo ; echo "Spin up mongod"
mongod --version
mkdir -p data/db
rm -rf data/db/*
mongod --port 30097 --dbpath data/db --logpath data/mongod.log --fork --wiredTigerCacheSizeGB .5
mongo --port 30097 _KEYHOLE_88800 --eval "db.setProfilingLevel(0, {slowms: 20})"

# Test version
echo ; echo "==> Test version (--version)"
go run keyhole.go --version

# Test quotes
echo ; echo "==> Test printing a quote (--quote)"
go run keyhole.go --quote
echo ; echo "==> Test printing all quotes (--quotes)"
go run keyhole.go --quotes

# Test Schema
echo ; echo "==> Test printing default schema (--schema)"
go run keyhole.go --schema
echo ; echo "==> Test printing schema from a template (--schema --file <file>)"
go run keyhole.go --schema --file examples/template.json

# Test Info
echo ; echo "==> Test printing cluster info (--uri <uri> --info)"
go run keyhole.go --uri mongodb://localhost:30097/KEYHOLEDB --info

# Test seed
echo ; echo "==> Test seeding default docs (--uri <uri> --seed)"
go run keyhole.go --uri mongodb://localhost:30097/KEYHOLEDB --seed
echo ; echo "==> Test seeding default docs after dropping collection (--uri <uri> --drop)"
go run keyhole.go --uri mongodb://localhost:30097/KEYHOLEDB --seed --drop
echo ; echo "==> Test seeding docs from a template (--uri <uri> --file <file>)"
go run keyhole.go --uri mongodb://localhost:30097/KEYHOLEDB --seed --file examples/template.json
echo ; echo "==> Test seeding docs from a template after dropping collection (--uri <uri> --file <file> --drop)"
go run keyhole.go --uri mongodb://localhost:30097/KEYHOLEDB --seed --file examples/template.json --drop

# Test load test
echo ; echo "==> Test load from a template (--uri <uri> --file <file>)"
go run keyhole.go --uri mongodb://localhost:30097/KEYHOLEDB --file examples/template.json --duration 2 \
    --tps 300 --conn 10 --simonly

go run keyhole.go --uri mongodb://localhost:30097/KEYHOLEDB --file examples/template.json --duration 3 \
    --tps 300 --conn 10 --tx examples/transactions.json

# Test loginfo
echo ; echo "==> Test printing performance stats from a log file (--loginfo <file>)"
go run keyhole.go --loginfo data/mongod.log

echo ; echo "Shutdown mongod"
mongo --port 30097 --eval 'db.getSisterDB("admin").shutdownServer()'
rm -rf data/db/*
