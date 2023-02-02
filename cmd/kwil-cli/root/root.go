package root

import (
	"fmt"
	"github.com/spf13/cobra"
	"kwil/cmd/kwil-cli/common"
	"kwil/cmd/kwil-cli/database"
	"kwil/cmd/kwil-cli/fund"
	initCli "kwil/cmd/kwil-cli/init"
	"path/filepath"
)

var rootCmd = &cobra.Command{
	Use:   "kwil-cli",
	Short: "kwil command line interface",
	Long:  "kwil-cli allows you to interact with the Kwil",
}

func Execute() error {
	rootCmd.AddCommand(
		fund.NewCmdFund(),
		//	configure.NewCmdConfigure(),
		database.NewCmdDatabase(),
		//	utils.NewCmdUtils(),
		initCli.NewCmdInit(),
	)
	//
	//common.BindGlobalFlags(rootCmd.PersistentFlags())
	//common.BindGlobalEnv(rootCmd.PersistentFlags())

	rootCmd.SilenceUsage = true
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(common.LoadConfig)
	defaultConfigPath := filepath.Join("$HOME", common.DefaultConfigDir, common.DefaultConfigName)
	rootCmd.PersistentFlags().StringVar(&common.ConfigFile, "config", "", fmt.Sprintf("config file (default is %s)", defaultConfigPath))
}
