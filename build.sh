#!/bin/bash

echo "Number of nodes to run (1-9): "
read x

mkdir -p log
cat /dev/null > log/log.txt

sudo docker build -t mxnode .

sudo docker network create --subnet=172.20.0.0/16 mxnetwork

sudo docker run --rm --net mxnetwork --ip 172.20.0.10 -v ${PWD}/log:/go/src/app/log -e NODE_NAME=node1 mxnode &>/dev/null
for (( i = 0; i < $((x-1)); i++))
do
  sleep 3
  sudo docker run --rm --net mxnetwork -v ${PWD}/log:/go/src/app/log -e NODE_NAME=node$((i+2)) -e CLUSTER_ADDRESS=172.20.0.10 mxnode &>/dev/null
done
echo DONE!
