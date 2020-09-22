#! /bin/bash
# Copyright 2018 Kuei-chun Chen. All rights reserved.

shutdownServer() {
    echo "Shutdown mongod" 
    mongo --quiet ${DATABASE_URI} ${TLS} --eval 'db.getSisterDB("admin").shutdownServer()' > /dev/null 2>&1
    rm -rf data/db data/mongod.log.*
    rm -f keyfile
    exit
}

validate() {
  if [ $? != 0 ]; then
      echo $1
      shutdownServer
  fi
}

echo ; echo "Spin up mongod"
mver=$(mongod --version|grep 'db version'|awk '{print $3}') 
mkdir -p data/db
rm -rf data/db/*
mongod --port 30097 --dbpath data/db --logpath data/mongod.log --fork --wiredTigerCacheSizeGB 1
validate "failed to start mongod"
rm -rf out/ 

mongo --quiet mongodb://localhost:30097/admin --eval 'db.createUser({ user: "user", pwd: "password", roles: [ "root" ] } )'
mongo --quiet mongodb://localhost:30097/admin --eval 'db.getSisterDB("admin").shutdownServer()' > /dev/null 2>&1
mkdir -p out/
openssl rand -base64 756 > out/keyfile
chmod 400 out/keyfile

if [[ -d "mdb/testdata/certs/" ]]; then
    export TLS="--tls"
    export TLS_STR="&tls=true"
    export TLS_MODE="--tlsMode requireTLS"
    export TLS_CLIENT="--tlsCAFile mdb/testdata/certs/ca.pem --tlsCertificateKeyFile mdb/testdata/certs/client.pem"
    export TLS_SERVER="--tlsCAFile mdb/testdata/certs/ca.pem --tlsCertificateKeyFile mdb/testdata/certs/server.pem"
fi

mongod --port 30097 --dbpath data/db --logpath data/mongod.log --fork \
    --wiredTigerCacheSizeGB 1 --keyFile out/keyfile --replSet rs ${TLS_MODE} ${TLS_SERVER}
    
echo "init replica set"
mongo --quiet mongodb://user:password@localhost:30097/admin --eval 'rs.initiate()' ${TLS} ${TLS_CLIENT} > /dev/null 2>&1
validate "init replica set"
sleep 2

export DATABASE_URI="${TLS_CLIENT} mongodb://user:password@localhost:30097/keyhole?authSource=admin&replicaSet=rs&readPreference=nearest${TLS_STR}"
mongo --quiet ${DATABASE_URI} ${TLS} --eval 'version()'

# Test version
echo ; echo "==> Test version (--version)"
go run main.go --version
validate "--version"

# Test Info
echo ; echo "==> Test printing cluster info (--info <uri>)"
go run main.go --info ${DATABASE_URI}
validate "--info <uri>"

# Test All Info
echo ; echo "==> Test printing cluster info (--allinfo <uri>)"
go run main.go --allinfo ${DATABASE_URI}
validate "--allinfo ${DATABASE_URI}"

# Test seed
echo ; echo "==> Test seeding default docs (--seed <uri>)"
go run main.go --seed ${DATABASE_URI}
validate "--seed ${DATABASE_URI}"

echo ; echo "==> Test seeding default docs after dropping collection (--seed --drop <uri>)"
go run main.go --seed --drop ${DATABASE_URI}
validate "--seed --drop ${DATABASE_URI}"

mongo ${DATABASE_URI} ${TLS} --eval "db.setProfilingLevel(0, {slowms: -1})"
validate "failed to set profiling level"
mongo ${DATABASE_URI} ${TLS} --eval 'db.vehicles.createIndex({color: 1})'
mongo ${DATABASE_URI} ${TLS} --eval 'db.vehicles.createIndex({color: 1, style: 1})'

echo ; echo "==> Test seeding docs from a template (--file <file> --collection <collection> <uri>)"
go run main.go --seed --file examples/template.json --collection template ${DATABASE_URI}
validate "--file <file> --collection <collection> <uri>"

echo ; echo "==> Test seeding docs from a template after dropping collection (--file <file> --collection <collection> --drop <uri>)"
go run main.go --seed --file examples/template.json --collection template --drop ${DATABASE_URI}
validate "--file <file> --collection <collection> --drop <uri>"

# Test Index
echo ; echo "==> Test printing cluster indexes (--index <uri>)"
go run main.go --index ${DATABASE_URI}
validate "--index <uri>"

# Test Schema
echo ; echo "==> Test printing schema from a template (--schema --collection <collection> <uri>)"
go run main.go --schema --collection vehicles ${DATABASE_URI}
validate "--schema --collection <collection> <uri>"

if [[ -d mdb/testdata/ ]]; then
    # Test loginfo
    echo ; echo "==> Test printing performance stats from a log file (--loginfo <file>)"
    go run main.go --loginfo mdb/testdata/mongod.text.log.gz
    go run main.go --loginfo mdb/testdata/mongod.json.log.gz
fi

if [[ "$mver" > "v3.4" ]]; then
    # Test Cardinality
    echo ; echo "==> Test printing number of distinct fileds values (--cardinality)"
    go run main.go --cardinality favorites ${DATABASE_URI}
    validate "--cardinality"

    # Test explain
    echo ; echo "==> Test explain (--explain)"
    go run main.go --explain mdb/testdata/vehicles.log ${DATABASE_URI}
    validate "--explain"
fi

# Test Create Index
echo ; echo "==> Test creating cluster indexes (--createIndex <index_info> <uri>)"
go run main.go --createIndex "out/$(hostname)-index.bson.gz" ${DATABASE_URI}
validate "--createIndex <index_info> <uri>"
rm -f "out/$(hostname)-index.bson.gz"
go run main.go --createIndex "out/$(hostname)_30097-stats.bson.gz" ${DATABASE_URI}
validate "--createIndex <index_info> <uri>"
rm -f "out/$(hostname)-stats.bson.gz"

if [ "$1" != "" ]; then
    # Test load test
    echo ; echo "==> Test load from a template (--file <file> <uri>)"
    go run main.go --yes --file examples/template.json --duration 1 \
        --tps 300 --conn 5 --simonly ${DATABASE_URI}
    validate "--file <file> <uri>"

    go run main.go --yes --file examples/template.json --duration 2 \
        --tps 300 --conn 5 --tx examples/transactions.json ${DATABASE_URI}
    rm -f keyhole_*.gz
    validate "--yes"
fi

# Test info Atlas
echo ; echo "==> Test printing clusters summary (--info <atlas_uri>)"
go run main.go --info "atlas://${ATLAS_AUTH}"

# Test loginfo Atlas
# echo ; echo "==> Test printing performance stats from a log file (--loginfo <atlas_uri>)"
# go run main.go --seed ${ATLAS_URL}
# go run main.go --loginfo "atlas://${ATLAS_AUTH}@${ATLAS_GROUP}/keyhole"
rm -f mongodb.log.*

shutdownServer
