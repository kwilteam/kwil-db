package main

import (
	"fmt"
	"os"

	root "github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds"
)

func main() {
	root := root.NewRootCmd()
	if err := root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	os.Exit(0)
}
