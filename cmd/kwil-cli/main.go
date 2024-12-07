package main

import (
	"os"

	// NOTE: if extensions are used to build a kwild with new transaction
	// payload types or serialization methods, the same extension packages that
	// register those types with core module packages would be imported here so
	// that the client can work with them too. While the client does is not
	// concerned with activation heights, it could need to use new functionality
	// introduced by the consensus extensions.

	root "github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
)

func main() {
	root := root.NewRootCmd()
	if err := root.Execute(); err != nil {
		config.PreRunPrintEffectiveConfig(root, nil) // only when --debug is set
		os.Exit(-1)
	}
	os.Exit(0)
}
