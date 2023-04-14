package app

import (
	"kwil/cmd/kwil-cli/cmds/configure"
	"kwil/cmd/kwil-cli/cmds/database"
	"kwil/cmd/kwil-cli/cmds/fund"
	"kwil/cmd/kwil-cli/cmds/system"
	"kwil/cmd/kwil-cli/cmds/utils"
	"kwil/cmd/kwil-cli/config"

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
		fund.NewCmdFund(),
		configure.NewCmdConfigure(),
		database.NewCmdDatabase(),
		utils.NewCmdUtils(),
		system.NewVersionCmd(),
	)

	return rootCmd.Execute()
}

func init() {
	config.BindGlobalFlags(rootCmd.PersistentFlags())
}
