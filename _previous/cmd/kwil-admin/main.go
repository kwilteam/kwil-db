package main

import (
	"fmt"
	"os"

	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds"
)

func main() {
	if err := cmds.NewRootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
