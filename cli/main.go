/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/kwilteam/kwil-db/cli/cmd"
	_ "github.com/kwilteam/kwil-db/cli/cmd/database"
	_ "github.com/kwilteam/kwil-db/cli/cmd/fund"
	_ "github.com/kwilteam/kwil-db/cli/cmd/set"
)

func main() {
	cmd.Execute()
}
