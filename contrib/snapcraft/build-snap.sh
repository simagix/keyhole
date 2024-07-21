#!/bin/bash
cp -r ./snap ../../
cd ../../
snapcraft "$@"
