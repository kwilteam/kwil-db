package migration

import (
	"github.com/spf13/cobra"

	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
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
		proposalStatusCmd(),
		genesisStateCmd(),
		networkStatusCmd(),
	)

	common.BindRPCFlags(migrationCmd)
	return migrationCmd
}
