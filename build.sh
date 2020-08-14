#! /bin/bash
# Copyright 2020 Kuei-chun Chen. All rights reserved.

if [[ "$(which go)" == "" ]]; then
    echo "go command not found"
    exit
fi

DEP=$(which dep)
if [[ "$DEP" == "" ]]; then
    echo "dep command not found"
    exit
fi

if [[ -d vendor ]]; then
    UPDATE="-update"
fi

mkdir -p dist
export ver=$(cat version)
export version="v${ver}-$(date "+%Y%m%d")"

if [[ "$1" == "all" ]]; then
  docker build  -f Dockerfile . -t simagix/keyhole
  id=$(docker create simagix/keyhole)
  docker cp $id:/dist - | tar x
  docker rmi -f $(docker images -f "dangling=true" -q)
else
  $DEP ensure $UPDATE
  if [ "$1" == "cross-platform"  ]; then
    env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$version" -o dist/keyhole-osx-x64 main.go
    env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$version" -o dist/keyhole-linux-x64 main.go
    env CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$version" -o dist/keyhole-win-x64.exe main.go
  else
    env CGO_ENABLED=0 go build -ldflags "-X main.version=$version" -o keyhole main.go
  fi
fi
