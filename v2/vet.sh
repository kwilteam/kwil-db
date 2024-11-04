#!/usr/bin/env bash

# be sure to "go work init ." first

set -e

echo "Tidying go.mod and go.work..."
go mod tidy
# ./scripts/tidy.sh

echo "Formating source..."
# go install golang.org/x/tools/cmd/goimports@v0.26.0
goimports -format-only -w .

echo "Running unit tests..."
go test -short ./...

echo "Compiling kwil node application..."
go build -o /dev/null .

echo "Linting all source..."
# go install -v github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golangci-lint run ./... -c .golangci.yml

echo "done"
