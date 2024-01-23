# Kwil DB

The database for Web3.

## Overview

Learn more about Kwil at [kwil.com](https://kwil.com).

## Build instructions

### Prerequisites

To build Kwil, you will need to install:

1. [Go](https://golang.org/doc/install)
2. [Protocol Buffers](https://developers.google.com/protocol-buffers/docs/gotutorial), with the `protoc` executable binary on your `PATH`.
3. [Taskfile](https://taskfile.dev/installation)
4. Protocol buffers go plugins and other command line tools.  The `tool` task will install the required versions of the tools into your `GOPATH`, so be sure to include `GOPATH/bin` on your `PATH`.

    ```shell
    task tools
    ```

### Build

Invoke `task` command to see all available tasks. The `build` task will compile `kwild`, `kwil-cli`, and `kwil-admin`. They will be generated in `.build/`:

```shell
task build
```

## Local deployment

You can start a single node network using the `kwild` binary built in the step above:

```shell
.build/kwild --autogen
```

For more information on running nodes, and how to run a multi-node network, refer to the Kwil [documentation](<https://docs.kwil.com/docs/node/quickstart>).

## Resetting local deployments

By default, `kwild` stores all data in `~/.kwild`. To reset the data on a deployment, remove the data directory while the node is stopped:

```shell
rm -r ~/.kwild
```
