#! /bin/bash
# Copyright 2020 Kuei-chun Chen. All rights reserved.
die() { echo "$*" 1>&2 ; exit 1; }
VERSION="v$(cat version)-$(date "+%Y%m%d")"
REPO=$(basename "$(dirname "$(pwd)")")/$(basename "$(pwd)")
LDFLAGS="-X main.version=$VERSION -X main.repo=$REPO"
TAG="simagix/keyhole"
[[ "$(which go)" = "" ]] && die "go command not found"
mkdir -p dist
if [[ "$1" == "docker" ]]; then
  docker rmi -f $(docker images -f "dangling=true" -q) > /dev/null 2>&1
  BR=$(git branch --show-current)
  if [[ "${BR}" == "master" ]]; then
    BR="latest"
  fi 
  docker build  -f Dockerfile -t ${TAG}:${BR} .
  id=$(docker create ${TAG}:${BR})
  docker cp $id:/dist - | tar vx
else
  if [ "$1" == "cross-platform"  ]; then
    env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/keyhole-osx-x64 main/keyhole.go
    env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/keyhole-linux-x64 main/keyhole.go
    env CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/keyhole-win-x64.exe main/keyhole.go
  else
    rm -f keyhole
    go build -ldflags "$LDFLAGS" main/keyhole.go
    if [[ -f keyhole ]]; then
      ./keyhole -version
    fi
  fi
fi
