package main

import (
	"os"

	"github.com/kwilteam/kwil-db/cmd"
	"github.com/kwilteam/kwil-db/cmd/kwild/root"
)

func main() {
	cmd.BinaryConfig.ProjectName = "kwild"
	if err := root.RootCmd().Execute(); err != nil {
		os.Exit(1) // cobra nicely prints the error already
	}
	os.Exit(0)
}
