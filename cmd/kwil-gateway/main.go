package main

import (
	"fmt"
	"kwil/cmd/kwil-gateway/server"
	"os"
)

func main() {
	if err := server.Start(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
