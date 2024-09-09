package cmds

import (
	"fmt"

	"github.com/kwilteam/kwil-db/cmd"
	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/common/version"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/account"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/configure"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/database"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/utils"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var longDesc = `Command line interface client for using %s.
	
` + "`" + `%s` + "`" + ` is a command line interface for interacting with %s. It can be used to deploy, update, and query databases.
	
` + "`" + `%s` + "`" + ` can be configured with a persistent configuration file. This file can be configured with the '%s configure' command.
` + "`" + `%s` + "`" + ` will look for a configuration file at ` + "`" + `$HOME/.kwil-cli/config.json` + "`" + `.`

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:               cmd.BinaryConfig.ClientCmd,
		Short:             fmt.Sprintf("Command line interface client for using %s.", cmd.BinaryConfig.ProjectName),
		Long:              fmt.Sprintf(longDesc, cmd.BinaryConfig.ProjectName, cmd.BinaryConfig.ClientUsage(), cmd.BinaryConfig.ProjectName, cmd.BinaryConfig.ClientUsage(), cmd.BinaryConfig.ClientUsage(), cmd.BinaryConfig.ClientUsage()),
		SilenceUsage:      true,
		DisableAutoGenTag: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// for backwards compatibility, we need to check if the deprecated flag is set.
			// If the new flag is set and the deprecated flag is not, we can proceed.
			// If both are set, we should return an error.
			if cmd.Flags().Changed("kwil-provider") {
				if cmd.Flags().Changed(config.GlobalProviderFlag) {
					return fmt.Errorf("cannot use both --provider and --kwil-provider flags")
				} else {
					viper.BindPFlag(config.GlobalProviderFlag, cmd.Flags().Lookup("kwil-provider"))
				}
			}

			return nil
		},
	}

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
