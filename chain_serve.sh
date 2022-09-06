#!/bin/sh
echo "Building package ..."
go build ./...

echo "Building kwild ..."
go build -o ./cmd/kwil-cosmos/cmd/kwild ./cmd/kwil-cosmos/cmd/kwild

echo "Serving Kwil Chain ..."
ignite chain serve $1 $2 $3 $4 $5 $6 $7 $8 $9