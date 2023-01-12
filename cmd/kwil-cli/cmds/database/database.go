package database

import (
	"kwil/cmd/kwil-cli/util"

	"github.com/spf13/cobra"
)

func NewCmdDatabase() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "database",
		Aliases: []string{"db"},
		Short:   "Database is a command that contains subcommands for interacting with databases",
		Long:    "",
	}

	cmd.AddCommand(
		viewDatabaseCmd(),
		deployCmd(),
		dropCmd(),
		listCmd(),
		executeCmd(),
	)
	util.BindKwilFlags(cmd.PersistentFlags())

	return cmd
}
