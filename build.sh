#!/bin/bash

echo "Input amount of nodes to run: "
read x

sudo docker build -t mxnode .

sudo docker network create --subnet=172.20.0.0/16 mxnetwork

mkdir -p log
rm ${PWD}/log/*.txt

touch ${PWD}/log/node1.txt
sudo docker run --rm --net mxnetwork --ip 172.20.0.10 -v ${PWD}/log/node1.txt:/go/src/app/log/log.txt -e NODE_NAME=node1 -e ADVERTISE_ADDRESS=172.20.0.10 -p 8080:8080 mxnode &
for (( i = 0; i < $((x-1)); i++))
do
  touch ${PWD}/log/node$((i+2)).txt
  sudo docker run --rm --net mxnetwork --ip 172.20.0.1$((i+1)) -v ${PWD}/log/node$((i+2)).txt:/go/src/app/log/log.txt -e NODE_NAME=node$((i+2)) -e ADVERTISE_ADDRESS=172.20.0.1$((i+1)) -e CLUSTER_ADDRESS=172.20.0.1$((i)) -p $((8081+i)):8080 mxnode &
done