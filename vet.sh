#!/usr/bin/env bash

task work

set -e

echo "Tidying go.mod and go.work..."
task tidy

echo "Formating source..."
# goimports: task tools
task fmt

echo "Running unit tests..."
task test:unit

echo "Compiling kwil node application..."
go build -o /dev/null ./cmd/kwild

echo "Linting all source..."
# install: task linter
task lint

echo "done"
