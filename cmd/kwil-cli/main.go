package main

import (
	"kwil/cmd/kwil-cli/app"
	"os"
)

func main() {
	if err := app.Execute(); err != nil {
		os.Exit(-1)
	}
}
