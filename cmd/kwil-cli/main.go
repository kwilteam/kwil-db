package main

import (
	"fmt"
	"os"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	root "github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds"
)

func main() {
	root := root.NewRootCmd()
	if err := root.Execute(); err != nil {
		err2 := display.PrintErr(root, err)
		if err2 != nil {
			fmt.Println(err2)
		}
		os.Exit(-1)
	}
}
