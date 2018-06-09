#! /bin/bash
# Copyright 2018 Kuei-chun Chen. All rights reserved.

echo ; echo "Spin up mongod"
mongod --version
mkdir -p data/db
mongod --port 33168 --dbpath data/db --logpath data/mongod.log --fork --wiredTigerCacheSizeGB .5

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
go run keyhole.go --uri mongodb://localhost:33168/keyhole --info

# Test seed
# echo ; echo "==> Test seeding default docs (--uri <uri> --seed)"
# go run keyhole.go --uri mongodb://localhost:33168/keyhole --seed
# echo ; echo "==> Test seeding default docs after dropping collection (--uri <uri> --drop)"
# go run keyhole.go --uri mongodb://localhost:33168/keyhole --seed --drop
echo ; echo "==> Test seeding docs from a template (--uri <uri> --file <file>)"
go run keyhole.go --uri mongodb://localhost:33168/keyhole --seed --file examples/template.json
echo ; echo "==> Test seeding docs from a template after dropping collection (--uri <uri> --file <file> --drop)"
go run keyhole.go --uri mongodb://localhost:33168/keyhole --seed --file examples/template.json --drop

# Test load test
echo ; echo "==> Test load from a template (--uri <uri> --file <file>)"
go run keyhole.go --uri mongodb://localhost:33168/keyhole --file examples/template.json --duration 2 \
    --tps 300 --conn 10 --tx examples/transactions.json

# Test loginfo
echo ; echo "==> Test printing performance stats from a log file (--loginfo <file>)"
go run keyhole.go --loginfo data/mongod.log

echo ; echo "Shutdown mongod"
mongo --port 33168 --eval 'db.getSisterDB("admin").shutdownServer()'
