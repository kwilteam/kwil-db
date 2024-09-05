package cmds

import (
	"fmt"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/common/version"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/key"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/migration"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/node"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/peers"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/setup"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/snapshot"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/utils"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/validators"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	return CustomRootCmd("kwil-admin", "Kwil")
}

func CustomRootCmd(usage, projectName string) *cobra.Command {
	desc := fmt.Sprintf("The %s node admin tool.", projectName)
	rootCmd := &cobra.Command{
		Use:               usage,
		Short:             desc,
		Long:              desc,
		SilenceUsage:      true,
		DisableAutoGenTag: true,
	}
	rootCmd.AddCommand(
		version.NewVersionCmd(),
		key.NewKeyCmd(),
		node.NewNodeCmd(),
		setup.NewSetupCmd(),
		validators.NewValidatorsCmd(),
		utils.NewUtilsCmd(),
		snapshot.NewSnapshotCmd(),
		peers.PeersCmd(),
		migration.NewMigrationCmd(),
	)

	display.BindOutputFormatFlag(rootCmd)
	display.BindSilenceFlag(rootCmd)

	return rootCmd
}
