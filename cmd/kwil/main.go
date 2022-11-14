package main

import (
	"fmt"
	"os"

	"kwil/x/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
