package main

import (
	"kwil/cmd/kwil-cli/cmds/configure"
	"kwil/cmd/kwil-cli/cmds/database"
	"kwil/cmd/kwil-cli/cmds/fund"
	"kwil/cmd/kwil-cli/cmds/utils"
	"kwil/cmd/kwil-cli/conf"

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
		//initCli.NewCmdInit(),
	)

	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(conf.LoadConfig)

	conf.BindGlobalFlags(rootCmd.PersistentFlags())
	conf.BindGlobalEnv(rootCmd.PersistentFlags())
}
