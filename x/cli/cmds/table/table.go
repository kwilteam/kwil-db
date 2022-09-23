package table

import (
	"github.com/spf13/cobra"
	"kwil/x/cli/util"
)

func NewCmdTable() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "table",
		Aliases: []string{"tbl"},
		Short:   "Table is a command that contains subcommands for interacting with database tables",
		Long:    "",
	}

	cmd.AddCommand(
		createTableCmd(),
		updateTableCmd(),
		dropTableCmd(),
		viewTableCmd(),
		listTablesCmd(),
	)

	util.BindKwilFlags(cmd.PersistentFlags())

	return cmd
}
