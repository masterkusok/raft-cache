# Raft Cache 

A distributed, highly-available, in-memory key-value store built with the Raft consensus algorithm in Go. This project is a practical implementation of a fault-tolerant state machine, perfect for learning and as a foundation for more complex distributed systems.

![Golang badge](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go)
![Raft-badge](https://img.shields.io/badge/Consensus-Raft-important)

## Overview
Raft Cache is a simplified version of  in-memory distributed key-value database. It ensures strong consistency across a cluster of machines by using the Raft protocol to replicate all state changes. If a majority of nodes are available, your data is safe and consistent, even if some nodes fail.

This is a pet-project - well-structured implementation that demonstrates the core concepts of distributed consensus.
## Key Features
- Raft Consensus: Implements leader election, log replication, and safety mechanisms as defined in the Raft paper.

- Strong Consistency: All reads and writes are handled by the cluster leader, guaranteeing linearizability.

- Simple HTTP API: Easy-to-use REST-like interface for GET, SET, and DELETE operations.

- Cluster Management: Built-in support for dynamically adding and removing nodes from the cluster.

- In-Memory Storage: Blazing fast read and write performance for the stored data.

## Getting Started
Prerequisites

-  Go 1.24+: Make sure you have Go installed on your system.

- Git: To clone the repository.

## Installation

#### Clone the repository:
``` bash
git clone https://github.com/masterkusok/raft-cache.git
cd raft-cache
```

#### Download the dependencies:
``` bash
go mod download
```

#### Build the project:
```bash
go build -o raft-cache cmd/main.go
```

## Running a Local Cluster

The easiest way to test a cluster is by running multiple nodes on your local machine.

#### Start the first node (the leader):
```bash

./raft-cache
```

This starts a node with ID node1, HTTP API on port 8000, Raft communication on port 8081, and initializes a cluster with itself.

#### Join second and third node:
``` bash
./raft-cache --local-id=node-2 --raft-addr=localhost:8082 --raft-dir=temp2/ --max-pool=3 --leader-addr=localhost:8082 --port=:8001 --leader-api-endpoint=http://localhost:8000

./raft-cache --local-id=node-3 --raft-addr=localhost:8083 --raft-dir=temp3/ --max-pool=3 --leader-addr=localhost:8081 --port=:8002 --leader-api-endpoint=http://localhost:8000
```

Those nodes will contact node1 and ask to join the cluster.

You now have a 3-node Raft cluster running! You can kill any single node, and the cluster will continue to operate normally, electing a new leader if necessary.

## API
You can use DB using simple http requests:

#### Set key value:
```
curl -X POST http://localhost:8000/api/v1/key/{key}/value/{value}
```

#### Get key value:
```
curl -X GET http://localhost:8000/api/v1/key/{key}
```

#### DELETE key:
```
curl -X DELETE http://localhost:8000/api/v1/key/{key}
```