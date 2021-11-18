# Distributed-Mutual-Exclusion
Mandatory Exercise 2 - Distributed Systems

## Running the application
```bash
# Build the image
docker build -t mxnode .

# Run multiple nodes
docker run -e NODE_NAME=node1 -e ADVERTISE_ADDRESS=<DockerIP> -p 8080:8080 mxnode
docker run -e NODE_NAME=node2 -e ADVERTISE_ADDRESS=172.17.0.4 -e CLUSTER_ADDRESS=172.17.0.3 -p 8081:8080 mxnode
docker run -e NODE_NAME=node3 -e ADVERTISE_ADDRESS=172.17.0.5 -e CLUSTER_ADDRESS=172.17.0.4 -p 8082:8080 mxnode
```

The token increases by one when it is passed to another node.