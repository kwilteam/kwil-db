package main

import (
	"kwil/internal/app/kwild"
	"os"
)

func main() {
	if err := kwild.Execute(); err != nil {
		os.Exit(-1)
	}
}
