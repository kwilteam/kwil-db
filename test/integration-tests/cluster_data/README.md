Put this folder under `deployments`.

## commands

* `./run start` to start the cluster
* `./run start build` to build docker image and start the cluster
* `./run stop` to stop the cluster

## configuration

.env file under kwil directory:
KWILD_DEPOSITS_CLIENT_CHAIN_RPC_URL=ws://{$IP}:8545
IP should be the host IP of the machine

### ganache

Everything is configured in `./ganache/docker-compose.yml`

* `wallet.mnemonic` this set the seed for available accounts

### kwil

* `./kwil/k1` contains k1 specific configuration.
* `./kwil/.env` is shared among all kwil instances
* `.env` file is used in docker-compose.yml to set environment variables
  * * if you set `THIS=3` in `.env`, you can get it by `os.Getenv("THIS")` in go code

### what concerns you

* `userPK` in `./kwil/eth_chain.go`, this is the private key of the user who will deploy database later
* `KWILD_DEPOSITS_CLIENT_CHAIN_RPC_URL` in `./kwil/.env`, this is the ETH rpc url you want to connect to
    * * change it to your LAN ip, and use `ws`(websocket) protocol

## how to interact with the cluster

* refer to docs on [kwil-cli](https://docs.kwil.com/cli/installation)
* run `task build:cli` in root directory to build binary, binary is localted at `ROOT/.build/`
* configure
  * * select one of the kwil instance, choose the GRPC port, this is the Kwil GRPC URL
  * * `userPK` in `./kwil/eth_chain.go` is your Private Key
  * * `KWILD_DEPOSITS_CLIENT_CHAIN_RPC_URL` in `./kwil/.env` is Client Chain RPC URL
