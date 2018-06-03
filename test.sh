#! /bin/bash
# Copyright 2018 Kuei-chun Chen. All rights reserved.

# Test version
go run keyhole.go --version

# Test Schema
go run keyhole.go --schema
go run keyhole.go --schema --file examples/seedkeys.json

# Test loginfo
go run keyhole.go --loginfo ~/ws/demo/mongod.log

# Test Info
go run keyhole.go --uri mongodb://localhost/ --info

# Test seed
go run keyhole.go --uri mongodb://localhost/ --seed
go run keyhole.go --uri mongodb://localhost/ --seed --drop
go run keyhole.go --uri mongodb://localhost/ --seed --file examples/seedkeys.json
go run keyhole.go --uri mongodb://localhost/ --seed --file examples/seedkeys.json --drop

# Test load test
go run keyhole.go --uri mongodb://localhost/ 
