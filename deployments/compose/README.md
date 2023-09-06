Put this folder under `deployments`.

## commands

* `./run start` to start the cluster
* `./run start build` to build docker image and start the cluster
* `./run stop` to stop the cluster and cleanup all the data

## configuration

* docker-compose.yml

* node config:
  * private_key stores the node private key used for p2p and signing messages
  * config.toml stores the app and chain configuration
  * abci/config/genesis.json contains the genesis file for setting up initial params

## how to interact with the cluster

* refer to docs on [kwil-cli](https://docs.kwil.com/cli/installation)
* run `task build:cli` in root directory to build binary, binary is localted at `ROOT/.build/`
* configure
  * select one of the kwil instance, choose the GRPC port, this is the Kwil GRPC URL

## COMETBFT Config

* Cometbft RPC Server: tcp://0.0.0.0:26657/
* Kwild RPC Server:  :50051
