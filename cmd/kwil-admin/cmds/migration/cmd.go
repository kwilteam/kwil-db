package migration

import (
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/spf13/cobra"
)

var migrationCmd = &cobra.Command{
	Use:   "migration",
	Short: "The `migration` command provides functions for managing migration proposals.",
	Long:  "The `migration` command provides functions for managing migration proposals.",
}

func NewMigrationCmd() *cobra.Command {
	migrationCmd.AddCommand(
		proposeCmd(),
		voteCmd(),
		listCmd(),
		statusCmd(),
		genesisStateCmd(),
	)

	common.BindRPCFlags(migrationCmd)
	return migrationCmd
}
