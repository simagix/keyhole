#! /bin/bash
# Copyright 2018 Kuei-chun Chen. All rights reserved.

export version="1.1"
mkdir -p build
env GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$version" -o build/keyhole-linux-x64 keyhole.go
env GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$version" -o build/keyhole-osx-x64 keyhole.go
env GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$version" -o build/keyhole-win-x64.exe keyhole.go

if [ "$1" == "docker" ]; then
    docker build -t simagix/keyhole .
fi
