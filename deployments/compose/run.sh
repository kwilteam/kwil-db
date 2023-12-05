#!/usr/bin/env bash

set -e

cleanup() {
  cp ./../../.build/kwild-darwin-arm64 kwild
  rm -rf kwil/testnet/
  rm ./kwild
}

bringup() {
  cp ./../../.build/kwild-darwin-arm64 kwild
  ./kwild utils testnet -o ./kwil/testnet -v 3 -n 0 --starting-ip-address 172.10.100.2 --populate-persistent-peers
  rm ./kwild
}

start() {
  # Build Kwild
  test $1 && task --taskfile ../../Taskfile.yml build:docker &&   task build || echo "skip build image"

  # start kwild
  printf "bringing up kwild services: \n"
  bringup

  docker compose -f kwil/docker-compose.yml up -d
}

stop() {
  docker-compose -f kwil/docker-compose.yml stop
  docker-compose -f kwil/docker-compose.yml rm -f
  cleanup
}

test $# -eq 0 && (echo Available funcs:;echo; declare -F | awk '{print $3}'; exit 1)

$@
