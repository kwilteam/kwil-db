/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package database

import (
	"github.com/kwilteam/kwil-db/cli/cmd"
	"github.com/spf13/cobra"
)

// databaseCmd represents the database command
var databaseCmd = &cobra.Command{
	Use:   "database",
	Short: "Database is a command that contains subcommands for interacting with databases",
	Long: `With the database command, you can create and modify databases.
	In the future, there will be more subcommands for interacting with databases.`,
}

func init() {
	cmd.RootCmd.AddCommand(databaseCmd)
}
