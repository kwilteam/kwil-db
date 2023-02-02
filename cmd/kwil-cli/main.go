package main

import (
	"kwil/cmd/kwil-cli/root"
	"os"
)

func main() {
	if err := root.Execute(); err != nil {
		//fmt.Println(err)
		os.Exit(-1)
	}
}
