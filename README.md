# Kwil DB

The database for Web3.

## Overview

Learn more about Kwil at [kwil.com](https://kwil.com).

## Build instructions

### Prerequisites

1. [Go](https://golang.org/doc/install)
2. [Protocol Buffers](https://developers.google.com/protocol-buffers/docs/gotutorial), with the `protoc` executable binary on your `PATH`.
3. [Taskfile](https://taskfile.dev/installation)
4. Protocol buffers go plugins and other command line tools.  The `tool` task will install the required versions of the tools into your `GOPATH`, so be sure to include `GOPATH/bin` on your `PATH`.

    ```shell
    task tools
    ```

### Build

Invoke `task` command to see all available tasks. The `build` task will compile both `kwild` and `kwil-cli`:

```shell
task build
```

## Local deployment

You can start a toy kwild and extension instance running in Docker containers managed by docker-compose using the confiration defined in `test/acceptance`.

This is not a production deployment.

### Prerequisites

1. [Docker](https://docs.docker.com/get-docker/)
2. [Docker Compose](https://docs.docker.com/compose/), included with Docker Desktop but not the standalone Docker Engine that provides the daemon and CLI client.
3. [Docker buildx](https://github.com/docker/buildx#linux-packages), for extended build capabilities with BuildKit. Also included with Docker Desktop.

### Run a local instance

```shell
# build local docker image
task build:docker
# run docker container
task dev:up
```
