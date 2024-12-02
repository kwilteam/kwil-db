#!/usr/bin/env bash

# be sure to "go work init ." first

set -e

echo "Tidying go.mod and go.work..."
(cd core; go mod tidy)
(cd parse; go mod tidy)
go mod tidy
# ./contrib/scripts/tidy.sh

echo "Formating source..."
# go install golang.org/x/tools/cmd/goimports@v0.26.0
goimports -format-only -w .

echo "Running unit tests..."
go test -short -count 1 ./...

echo "Compiling kwil node application..."
go build -o /dev/null ./cmd/kwild

echo "Linting all source..."
# go install -v github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golangci-lint run ./... -c .golangci.yml

echo "done"
