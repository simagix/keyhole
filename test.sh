#! /bin/bash
# Copyright 2018 Kuei-chun Chen. All rights reserved.

mongod --version
mkdir -p data/db
mongod --port 33168 --dbpath data/db --logpath data/mongod.log --fork

# Test version
go run keyhole.go --version

# Test quotes
go run keyhole.go --quote
go run keyhole.go --quotes

# Test Schema
go run keyhole.go --schema
go run keyhole.go --schema --file examples/seedkeys.json

# Test Info
go run keyhole.go --uri mongodb://localhost:33168/ --info

# Test seed
go run keyhole.go --uri mongodb://localhost:33168/ --seed
go run keyhole.go --uri mongodb://localhost:33168/ --seed --drop
go run keyhole.go --uri mongodb://localhost:33168/ --seed --file examples/seedkeys.json
go run keyhole.go --uri mongodb://localhost:33168/ --seed --file examples/seedkeys.json --drop

# Test load test
go run keyhole.go --uri mongodb://localhost:33168/ --file examples/template.json

# Test loginfo
go run keyhole.go --loginfo data/mongod.log

mongo --port 33168 --eval 'db.getSisterDB("admin").shutdownServer()'

