package main

import (
	"github.com/kwilteam/kwil-db/internal/sql/postgres"
)

func main() {
	doc, err := postgres.ParsePaths("data/test.hcl")
	if err != nil {
		panic(err)
	}
	_ = doc
}
