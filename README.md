# KwilDB

The database for Web3.

![banner](./assets/kwil-banner.jpeg) 

![Github-License](https://img.shields.io/badge/license-Apache%202.0-green) ![Release](https://img.shields.io/github/v/release/kwilteam/kwil-db) ![Build-Status](https://github.com/kwilteam/kwil-db/actions/workflows/ci.yaml/badge.svg) ![Go-Version](https://img.shields.io/github/go-mod/go-version/kwilteam/kwil-db) ![GoDoc](https://godoc.org/github.com/kwilteam/kwil-db?status.svg)
![Go-Report-Card](https://goreportcard.com/badge/github.com/kwilteam/kwil-db) ![Discord](https://img.shields.io/discord/819855804554543114?logo=discord)

KwilDB is the node software for Kwil Networks. Built with [PostgreSQL](https://www.postgresql.org/) and [CometBFT](https://github.com/cometbft/cometbft), KwilDB enables byzantine fault tolerant networks to be built on top of relational databases.

## Overview

To learn more about high-level Kwil concepts, refer to the [Kwil documentation](https://docs.kwil.com/docs/concepts).

To test deploying and using a Kuneiform schema (Kwil's smart contract language) on the Kwil testnet, refer to the [Kwil testnet tutorial](https://docs.kwil.com/docs/getting-started/create-a-database).

For more information on kwil-db, check out the [Kwil node documentation](https://docs.kwil.com/docs/node/quickstart).

## Quickstart

### Build Instructions

#### Prerequisites

To build Kwil, you will need to install:

1. [Go](https://golang.org/doc/install)
2. [Protocol Buffers](https://protobuf.dev/downloads/) (optional), with the `protoc` executable binary on your `PATH`.
3. [Taskfile](https://taskfile.dev/installation)
4. [Docker](https://docs.docker.com/get-docker/) to run a PostgreSQL database.
5. Miscellaneous go plugins and other command line tools. The `tools` task will install the required versions of the tools into your `GOPATH`, so be sure to include `GOPATH/bin` on your `PATH`.

    ```shell
    task tools
    ```

#### Build

Invoke `task` command to see all available tasks. The `build` task will compile `kwild`, `kwil-cli`, and `kwil-admin` binaries. They will be generated in `.build/`:

```shell
task build
```

### Local Deployment

#### Start PostgreSQL

Each Kwil node requires a PostgreSQL instance to run. You can start a PostgreSQL database using Docker Compose:

```shell
docker run -p 5432:5432 -v kwil-pg-demo:/var/lib/postgresql/data \
    --shm-size 256m -e "POSTGRES_HOST_AUTH_METHOD=trust" \
    --name kwil-pg-demo kwildb/postgres:latest
```

### Start kwild

You can start a single node network using the `kwild` binary built in the step above:

```shell
.build/kwild --autogen
```

For more information on running nodes, and how to run a multi-node network, refer to the Kwil [documentation](https://docs.kwil.com/docs/node/quickstart).

### Resetting local deployments

By default, `kwild` stores all data in `~/.kwild`. To reset the data on a deployment, remove the data directory while the node is stopped:

```shell
rm -r ~/.kwild
```

## Extensions

Kwil offers an extension system that allows you to extend the functionality of your network (e.g. building network oracles, customizing authentication, running deterministic compute, etc.). To learn more about the types of extensions and how to build them, refer to the extensions directory [README](extensions/README.md).

## Contributing

We welcome contributions to kwil-db. To contribute, please read our [contributing guidelines](CONTRIBUTING.md).

## License

The kwil-db repository (i.e. everything outside of the `core` directory) is licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for more details.

The kwil Go SDK (i.e. everything inside of the `core` directory) is licensed under the MIT License. See [core/LICENSE.md](core/LICENSE.md) for more details.
