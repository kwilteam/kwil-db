#!/usr/bin/env bash

set -e

cleanup() {
  cp ./../../.build/kwild-darwin-arm64 kwild
  ./kwild utils unsafe_reset_all --root_dir ./kwil/k1/node0
  ./kwild utils unsafe_reset_all --root_dir ./kwil/k2/node1
  ./kwild utils unsafe_reset_all --root_dir ./kwil/k3/node2
  rm ./kwild
}

start() {
  # Build Kwild
  test $1 && task --taskfile ../../Taskfile.yml build:docker -- shell &&   task build || echo "skip build image"

  # start kwild
  printf "bringing up kwild services: \n"
  docker compose -f kwil/docker-compose.yml up -d
}

stop() {
  docker-compose -f kwil/docker-compose.yml stop
  docker-compose -f kwil/docker-compose.yml rm -f
  cleanup
}

test $# -eq 0 && (echo Available funcs:;echo; declare -F | awk '{print $3}'; exit 1)

$@
