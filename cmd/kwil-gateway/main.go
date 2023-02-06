package main

import (
	"fmt"
	"kwil/internal/app/kwil-gateway"
	"os"
)

func main() {
	if err := kwil_gateway.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
