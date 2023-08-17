package main

import (
	"os"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/app"
)

func main() {
	if err := app.Execute(); err != nil {
		os.Exit(-1)
	}
}
