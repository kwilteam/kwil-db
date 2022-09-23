package main

import (
	"crypto/sha512"
	"encoding/json"
	"fmt"
)

func main() {
	y := yeuh{
		Name:    "hello",
		Version: "world",
	}

	// make sha384 hash
	hash := sha512.New384()

	b, _ := y.Marshal()
	fmt.Println(string(b))
	hash.Write(b)

	fmt.Printf("%x", hash.Sum(nil))
	fmt.Println()
}

type yeuh struct {
	Name    string
	Version string
}

func (y *yeuh) Marshal() ([]byte, error) {
	return json.Marshal(y)
}
