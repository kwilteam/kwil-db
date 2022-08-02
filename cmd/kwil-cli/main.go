package main

import (
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/app"
	"os"
)

func main() {
	if err := app.Execute(); err != nil {
		os.Exit(-1)
	}
}
