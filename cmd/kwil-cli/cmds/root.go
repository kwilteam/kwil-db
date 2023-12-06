package cmds

import (
	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/common/version"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/account"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/configure"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/database"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/utils"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	config.BindGlobalFlags(rootCmd.PersistentFlags())
	display.BindOutputFormatFlag(rootCmd)
	display.BindSilenceFlag(rootCmd)
	common.BindAssumeYesFlag(rootCmd)

	rootCmd.AddCommand(
		account.NewCmdAccount(),
		configure.NewCmdConfigure(),
		database.NewCmdDatabase(),
		utils.NewCmdUtils(),
		version.NewVersionCmd(),
	)

	return rootCmd
}

var rootCmd = &cobra.Command{
	Use:   "kwil-cli",
	Short: "Command line interface for using Kwil databases.",
	Long: `Command line interface for using Kwil databases.

The Kwil CLI is a command line interface for interacting with Kwil databases.  It can be used to deploy, update, and query databases.  It can also be used to generate documentation for Kwil databases.

The Kwil CLI can be configured with a persistent configuration file.  This file can be configured with the 'kwil-cli configure' command.  The Kwil CLI will look for a configuration file at ` + "`" + `$HOME/.kwil-cli/config.json` + "`" + `.
	`,
	SilenceUsage:      true,
	DisableAutoGenTag: true,
}
