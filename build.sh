#!/bin/bash
echo "Input amount of nodes to run: "
read x

touch log1 log2

sudo docker build -t mxnode .

gateway=$(sudo docker network inspect bridge | grep Gateway | awk '{print substr($2,2,length($2)-3)}')
ip=$(sudo docker network inspect bridge | grep Gateway | awk '{print substr($2,length($2)-1,1)}')

mkdir -p ${PWD}/log

sudo docker run --rm -v ${PWD}/log:/go/src/app/log -e NODE_NAME=node1 -e ADVERTISE_ADDRESS=${gateway}$((ip+1)) -p 8080:8080 mxnode &
for i in {0..${x}}
do
  sudo docker run --rm -v ${PWD}/log:/go/src/app/log -e NODE_NAME=node${i+2} -e ADVERTISE_ADDRESS=${gateway}$((ip+i+1)) -e CLUSTER_ADDRESS=${gateway}$((ip+i)) -p $((8080+i)):8080 mxnode &
done