package main

import (
	"os"

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

func execute() error {
	rootCmd.AddCommand(
		configure.NewCmdConfigure(),
		database.NewCmdDatabase(),
		utils.NewCmdUtils(),
		system.NewVersionCmd(),
	)

	return rootCmd.Execute()
}

func main() {
	config.BindGlobalFlags(rootCmd.PersistentFlags())

	if err := execute(); err != nil {
		os.Exit(-1)
	}
}
