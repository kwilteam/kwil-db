package database

import (
	"github.com/spf13/cobra"
	"kwil/x/cli/util"
)

func NewCmdDatabase() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "database",
		Aliases: []string{"db"},
		Short:   "Database is a command that contains subcommands for interacting with databases",
		Long:    "",
	}

	cmd.AddCommand(
		createDatabaseCmd(),
		updateDatabaseCmd(),
		dropDatabaseCmd(),
		viewDatabaseCmd(),
		listDatabaseCmd(),
	)

	util.BindKwilFlags(cmd.PersistentFlags())

	return cmd
}
