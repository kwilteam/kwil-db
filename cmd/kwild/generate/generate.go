package main

import (
	"flag"

	"github.com/kwilteam/kwil-db/app"
	"github.com/kwilteam/kwil-db/app/shared/generate"
)

var (
	out string
)

func main() {
	flag.StringVar(&out, "out", "./dist", "output directory")

	flag.Parse()

	err := generate.WriteDocs(app.RootCmd(), out)
	if err != nil {
		panic(err)
	}
}
