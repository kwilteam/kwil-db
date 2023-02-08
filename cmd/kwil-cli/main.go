package main

import (
	"kwil/internal/app/kcli"
	"os"
)

func main() {
	if err := kcli.Execute(); err != nil {
		os.Exit(-1)
	}
}
