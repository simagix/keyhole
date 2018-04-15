#! /bin/bash
mkdir -p build
env GOOS=linux GOARCH=amd64 go build -o build/keyhole-linux-x64 keyhole.go
env GOOS=windows GOARCH=amd64 go build -o build/keyhole-win-x64.exe keyhole.go
env GOOS=darwin GOARCH=amd64 go build -o build/keyhole-osx-x64 keyhole.go
