package main

import (
	"os"

	"github.com/kwilteam/kwil-db/cmd/kwild/root"
)

func main() {
	if err := root.RootCmd().Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
