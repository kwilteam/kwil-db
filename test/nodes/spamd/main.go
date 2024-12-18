package main

import (
	"fmt"
	"os"

	"github.com/kwilteam/kwil-db/app"
)

func main() {
	if err := app.RootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}
