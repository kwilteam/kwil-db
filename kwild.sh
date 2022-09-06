#!/bin/sh

echo "Building kwild ..."
go build -o ./cmd/kwil-cosmos/cmd/kwild ./cmd/kwil-cosmos/cmd/kwild

echo "Running kwild  ..."
./cmd/kwil-cosmos/cmd/kwild/kwild $1 $2 $3 $4 $5 $6 $7 $8 $9 --home $HOME/.kwildb/chain
