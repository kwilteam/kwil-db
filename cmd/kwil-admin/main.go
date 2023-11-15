package main

import (
	"fmt"
	"os"

	"github.com/kwilteam/kwil-db/cmd/internal/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds"
)

func main() {
	root := cmds.NewRootCmd()
	if err := root.Execute(); err != nil {
		err2 := display.PrintErr(root, err)
		if err2 != nil {
			fmt.Println(err2)
		}
		os.Exit(-1)
	}
}
