#! /bin/bash
export version="master-$(date "+%Y%m%d.%s")"
mkdir -p build
env GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$version" -o build/keyhole-linux-x64 keyhole.go
env GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$version" -o build/keyhole-osx-x64 keyhole.go
env GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$version" -o build/keyhole-win-x64.exe keyhole.go
