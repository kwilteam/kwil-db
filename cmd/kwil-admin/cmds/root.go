package cmds

import (
	"fmt"

	"github.com/kwilteam/kwil-db/cmd"
	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/common/version"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/key"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/migration"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/node"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/setup"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/snapshot"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/utils"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/validators"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/whitelist"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	desc := fmt.Sprintf("The %s node admin tool.", cmd.BinaryConfig.ProjectName)
	rootCmd := &cobra.Command{
		Use:               cmd.BinaryConfig.AdminCmd,
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
		whitelist.WhitelistCmd(),
		migration.NewMigrationCmd(),
	)

	display.BindOutputFormatFlag(rootCmd)
	display.BindSilenceFlag(rootCmd)

	return rootCmd
}
