package main

import (
	"fmt"
	_ "net/http/pprof"
	"os"

	"github.com/kwilteam/kwil-db/cmd/kwild/root"

	_ "github.com/kwilteam/kwil-db/extensions" // a base location where all extensions can be registered
	_ "github.com/kwilteam/kwil-db/extensions/auth"
)

func main() {
	if err := root.RootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}
