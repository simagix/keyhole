#! /bin/bash
# Copyright 2018 Kuei-chun Chen. All rights reserved.

# dep init

DEP=`which dep`

if [ "$DEP" == "" ]; then
    echo "dep command not found"
    exit
fi

if [ -d vendor ]; then
    UPDATE="-update"
fi

$DEP ensure $UPDATE
mkdir -p build
export ver="2.3.5"
export version="v${ver}-$(date "+%Y%m%d")"
env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$version" -o build/keyhole-osx-x64 keyhole.go

if [ "$1" == "all"  ]; then
  env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$version" -o build/keyhole-linux-x64 keyhole.go
  env CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$version" -o build/keyhole-win-x64.exe keyhole.go
fi

if [ "$1" == "docker"  ]; then
  docker build . -t simagix/keyhole
  id=$(docker create simagix/keyhole)
  docker cp $id:/build - | tar x
  docker rmi -f $(docker images -f "dangling=true" -q)
fi
