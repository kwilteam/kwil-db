package table

import (
	"github.com/kwilteam/kwil-db/internal/cli/util"
	"github.com/spf13/cobra"
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
