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

	deploy := deployCmd()
	deploy.Flags().StringP("path", "p", "", "Path to the database definition file")
	deploy.MarkFlagRequired("path")

	cmd.AddCommand(
		viewDatabaseCmd(),
		deploy,
		dropCmd(),
	)

	util.BindKwilFlags(cmd.PersistentFlags())

	return cmd
}
