package main

import (
	"os"

	"github.com/kwilteam/kwil-db/cmd/common"
	"github.com/kwilteam/kwil-db/cmd/kwild/root"
)

func main() {
	common.BinaryConfig.ProjectName = "kwild"
	if err := root.RootCmd().Execute(); err != nil {
		os.Exit(1) // cobra nicely prints the error already
	}
	os.Exit(0)
}
