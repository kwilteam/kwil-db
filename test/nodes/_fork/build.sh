#!/usr/bin/env bash

set -eu

task build:docker

export DOCKER_BUILDKIT=1
IMAGE="kwild-forker"

docker build ../../.. -t "${IMAGE}:latest" -f "dockerfile" # --no-cache
