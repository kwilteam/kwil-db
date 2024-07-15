package migration

import (
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/spf13/cobra"
)

var migrationCmd = &cobra.Command{
	Use:   "migration",
	Short: "The `migration` command provides functions for triggering and voting on migration transactions.",
	Long:  "The `migration` command provides functions for triggering and voting on migration transactions.",
}

func NewMigrationCmd() *cobra.Command {
	migrationCmd.AddCommand(
		triggerCmd(),
		voteCmd(),
		listCmd(),
		genesisStateCmd(),
	)

	common.BindRPCFlags(migrationCmd)
	return migrationCmd
}
