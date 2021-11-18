#!/bin/bash

echo "Input amount of nodes to run: "
read x

sudo docker build -t mxnode .

gateway=$(sudo docker network inspect bridge | grep Gateway | awk '{print substr($2,2,length($2)-3)}')
ip=$(sudo docker network inspect bridge | grep Gateway | awk '{print substr($2,length($2)-1,1)}')
((ip+=3))

mkdir -p log
rm ${PWD}/log/*.txt

touch ${PWD}/log/node1.txt
sudo docker run --rm -v ${PWD}/log/node1.txt:/go/src/app/log/log.txt -e NODE_NAME=node1 -e ADVERTISE_ADDRESS=${gateway}$((ip+1)) -p 8080:8080 mxnode & > /dev/null
for i in $(eval echo {0..$((x-2))})
do
  touch ${PWD}/log/node${i+2}.txt
  sudo docker run --rm -v ${PWD}/log/node${i+2}.txt:/go/src/app/log/log.txt -e NODE_NAME=node${i+2} -e ADVERTISE_ADDRESS=${gateway}$((ip+i+2)) -e CLUSTER_ADDRESS=${gateway}$((ip+i+1)) -p $((8080+i+1)):8080 mxnode & > /dev/null
done