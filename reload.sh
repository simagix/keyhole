#! /bin/bash
dir=$1
span=$2
if [ "$dir" == "" ]; then
    dir="/Users/kenchen/Downloads/diagnostic.data.trimmed/"
fi
if [ "$span" == "" ]; then
    span=10
fi

data="{\"dir\": \"$dir\", \"span\": $span}"
curl -XPOST http://localhost:5408/grafana/dir -d "$data"

