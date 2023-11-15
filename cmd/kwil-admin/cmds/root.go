package cmds

import (
	"github.com/kwilteam/kwil-db/cmd/internal/display"
	"github.com/kwilteam/kwil-db/cmd/internal/version"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/key"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/node"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/setup"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/validators"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	rootCmd.AddCommand(
		version.NewVersionCmd(),
		key.NewKeyCmd(),
		node.NewNodeCmd(),
		setup.NewSetupCmd(),
		validators.NewValidatorsCmd(),
	)

	display.BindOutputFormatFlag(rootCmd)
	display.BindSilenceFlag(rootCmd)

	return rootCmd
}

var rootCmd = &cobra.Command{
	Use:               "kwil-admin",
	Short:             "The Kwil node admin tool.",
	Long:              `The Kwil node admin tool.`,
	SilenceUsage:      true,
	DisableAutoGenTag: true,
}
