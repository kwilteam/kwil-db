#!/usr/bin/env bash

set -eu

task build:docker

export DOCKER_BUILDKIT=1
IMAGE="kwild-spammer"

docker build ../../.. -t "${IMAGE}:latest" -f "dockerfile" # --no-cache
