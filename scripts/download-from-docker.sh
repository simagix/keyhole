#!/bin/bash
docker rmi -f simagix/keyhole
id=$(docker create simagix/keyhole)
docker cp $id:/dist - | tar vx
docker rm $id
