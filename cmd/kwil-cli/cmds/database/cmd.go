package database

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:     "database",
		Aliases: []string{"db"},
		Short:   "manage databases",
		Long:    "Database is a command that contains subcommands for interacting with databases",
	}
)

func NewCmdDatabase() *cobra.Command {
	rootCmd.AddCommand(
		deployCmd(),
		dropCmd(),
		readSchemaCmd(),
		executeCmd(),
		listCmd(),
		batchCmd(),
		queryCmd(),
		callCmd(),
	)

	return rootCmd
}
