#!/usr/bin/env bash

set -e

start() {
  # Build Kwild
  test $1 && task --taskfile ../../Taskfile.yml build:docker -- shell || echo "skip build image"
  #task --taskfile ../../Taskfile.yml docker:kwild -- shell

  # start kwild

  printf "bringing up kwild\n"

  cp -r ./kwil/k1/node0-cpy/ ./kwil/k1/node0
  cp -r ./kwil/k2/node1-cpy/ ./kwil/k2/node1
  cp -r ./kwil/k3/node2-cpy/ ./kwil/k3/node2


  docker compose -f kwil/docker-compose.yml up -d
}

stop() {
  docker-compose -f kwil/docker-compose.yml stop
  docker-compose -f kwil/docker-compose.yml rm -f

  docker-compose -f ganache/docker-compose.yml stop
  docker-compose -f ganache/docker-compose.yml rm -f

  rm -rf ./kwil/k1/node0
  rm -rf ./kwil/k2/node1
  rm -rf ./kwil/k3/node2
}

test $# -eq 0 && (echo Available funcs:;echo; declare -F | awk '{print $3}'; exit 1)

$@
