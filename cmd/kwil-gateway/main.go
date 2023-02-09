package main

import (
	"fmt"
	"kwil/internal/app/kgw"
	"os"
)

func main() {
	if err := kgw.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
