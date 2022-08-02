package main

import (
	"github.com/kwilteam/kwil-db/internal/app/kwild"
	"os"
)

func main() {
	if err := kwild.Execute(); err != nil {
		os.Exit(-1)
	}
}
