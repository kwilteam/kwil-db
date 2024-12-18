package migration

import (
	"github.com/spf13/cobra"

	"github.com/kwilteam/kwil-db/app/rpc"
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

	rpc.BindRPCFlags(migrationCmd)
	return migrationCmd
}
