#! /bin/bash
# Copyright 2020 Kuei-chun Chen. All rights reserved.
die() { echo "$*" 1>&2 ; exit 1; }
VERSION="v$(cat version)-$(date "+%Y%m%d")"
REPO=$(basename "$(dirname "$(pwd)")")/$(basename "$(pwd)")
LDFLAGS="-X main.version=$VERSION -X main.repo=$REPO"
[[ "$(which go)" = "" ]] && die "go command not found"
[[ "$GOPATH" = "" ]] && die "GOPATH not set"
[[ "${GOPATH}/src/github.com/$REPO" != "$(pwd)" ]] && die "building keyhole should be under ${GOPATH}/src/github.com/$REPO"
mkdir -p dist
if [[ "$1" == "all" ]]; then
  docker rmi -f $(docker images -f "dangling=true" -q) > /dev/null 2>&1
  docker build  -f Dockerfile . -t $REPO
  id=$(docker create $REPO)
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
