package migration

import (
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/spf13/cobra"
)

var migrationCmd = &cobra.Command{
	Use:   "migrate",
	Short: "The `migrate` command provides functions for managing migration proposals.",
	Long:  "The `migrate` command provides functions for managing migration proposals.",
}

func NewMigrationCmd() *cobra.Command {
	migrationCmd.AddCommand(
		proposeCmd(),
		approveCmd(),
		listCmd(),
		statusCmd(),
		genesisStateCmd(),
	)

	common.BindRPCFlags(migrationCmd)
	return migrationCmd
}
