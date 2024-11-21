package main

import (
	"flag"

	"github.com/kwilteam/kwil-db/app/common/generate"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds"
)

var (
	out string
)

func main() {
	flag.StringVar(&out, "out", "./dist", "output directory")

	flag.Parse()

	err := generate.WriteDocs(cmds.NewRootCmd(), out)
	if err != nil {
		panic(err)
	}
}
