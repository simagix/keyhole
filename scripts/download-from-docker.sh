#!/bin/bash
docker rmi -f simagix/keyhole
docker rmi simagix/keyhole
id=$(docker create simagix/keyhole)
docker cp $id:/dist - | tar x
ls -l dist
