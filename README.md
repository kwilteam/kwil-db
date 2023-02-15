# Kwil db
The database for Web3.

## Overview
Learn more about Kwil at [kwil.com](https://kwil.com)

## Build instructions

Check `task list` to see all available tasks.

### Prerequisites
1. [Go](https://golang.org/doc/install)
2. [Protocol Buffers](https://developers.google.com/protocol-buffers/docs/gotutorial)
3. Protocol plugis:
```
  go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway
  go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2
  go install google.golang.org/protobuf/cmd/protoc-gen-go
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
```

### Build
1. Update submodules:
```
  git submodule update --init --recursive
```
2. Build the binary:
```
  task build
```

## Kubernetes deployment
This is a much easier way to deploy the kwil service.

### Prerequisites
1. [Docker](https://docs.docker.com/get-docker/)
2. [Kubernetes](https://kubernetes.io/docs/setup/)
3. [Helm](https://helm.sh/docs/intro/install/)

### Deploy
```
  task k8s:kwil
```