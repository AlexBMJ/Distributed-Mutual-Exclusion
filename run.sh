#!/bin/bash

echo "Number of nodes to run (1-9): "
read x

mkdir -p log
cat /dev/null > log/log.txt

sudo docker build -t mxnode .

sudo docker network create --subnet=172.20.0.0/16 mxnetwork

sudo docker run -d --rm --net mxnetwork --ip 172.20.0.10 -v ${PWD}/log:/go/src/app/log -e NODE_NAME=node1 mxnode
for (( i = 0; i < $((x-1)); i++))
do
  sleep 2
  sudo docker run -d --rm --net mxnetwork -v ${PWD}/log:/go/src/app/log -e NODE_NAME=node$((i+2)) -e CLUSTER_ADDRESS=172.20.0.10 mxnode
done
echo DONE!
sleep 1
tail -n +0 -f log/log.txt
