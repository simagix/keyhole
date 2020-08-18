#! /bin/bash
export GOPATH=$(pwd)/go
mkdir -p $GOPATH/src/github.com/simagix
cd $GOPATH/src/github.com/simagix
rm -rf keyhole
git clone --depth 1 https://github.com/simagix/keyhole.git
cd keyhole
./build.sh
./keyhole -version
