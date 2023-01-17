package database

import (
	"kwil/cmd/kwil-cli/common"

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
		viewDatabaseCmd(),
		executeCmd(),
		listCmd(),
	)

	common.BindKwilFlags(rootCmd)
	common.BindKwilEnv(rootCmd)
	common.BindChainFlags(rootCmd)
	common.BindChainEnv(rootCmd)

	return rootCmd
}
