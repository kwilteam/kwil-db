package main

import (
	"os"

	"github.com/kwilteam/kwil-db/internal/app/kwild"
)

func main() {
	if err := kwild.Execute(); err != nil {
		os.Exit(-1)
	}
}
