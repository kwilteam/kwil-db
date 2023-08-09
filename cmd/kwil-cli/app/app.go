package app

import (
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/configure"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/database"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/system"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/utils"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:               "kwil-cli",
	Short:             "kwil command line interface",
	Long:              "kwil-cli allows you to interact with the Kwil",
	SilenceUsage:      true,
	DisableAutoGenTag: true,
}

func Execute() error {
	rootCmd.AddCommand(
		configure.NewCmdConfigure(),
		database.NewCmdDatabase(),
		utils.NewCmdUtils(),
		system.NewVersionCmd(),
	)

	err := rootCmd.Execute()
	if err != nil {
		return err
	}

	return nil
}

func init() {
	config.BindGlobalFlags(rootCmd.PersistentFlags())
}
