# Kwil db
The database for Web3.

## Overview
Learn more about Kwil at [kwil.com](https://kwil.com)

## Build instructions

### Prerequisites
1. [Go](https://golang.org/doc/install)
2. [Protocol Buffers](https://developers.google.com/protocol-buffers/docs/gotutorial)
3. [Taskfile](https://taskfile.dev/installation)
3. Protocol buffers go plugins:
```shell
  task install:deps
```

### Build
Invoke `task` command to see all available tasks.

```shell
  task build
```

## Local deployment

### Prerequisites
1. [Docker](https://docs.docker.com/get-docker/)

### Run a local instance
```shell
  # build local docker image
  task docker:kwild
  # run docker container
  task dev:up
```
