/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	cmd "github.com/kwilteam/kwil-db/internal/cli/commands"
	_ "github.com/kwilteam/kwil-db/internal/cli/commands/database"
	_ "github.com/kwilteam/kwil-db/internal/cli/commands/fund"
	_ "github.com/kwilteam/kwil-db/internal/cli/commands/set"
)

func main() {
	cmd.Execute()
}
