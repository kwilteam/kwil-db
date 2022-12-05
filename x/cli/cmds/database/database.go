package database

import (
	"kwil/x/cli/util"

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
	)

	util.BindKwilFlags(cmd.PersistentFlags())

	return cmd
}
