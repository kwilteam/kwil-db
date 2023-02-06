package kcli

import (
	"fmt"
	"github.com/spf13/cobra"
	"kwil/internal/app/kcli/config"
	"kwil/internal/app/kcli/database"
	"kwil/internal/app/kcli/fund"
	"kwil/internal/app/kcli/utils"
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
		utils.NewCmdUtils(),
		//initCli.NewCmdInit(),
	)

	rootCmd.SilenceUsage = true
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(config.LoadConfig)

	defaultConfigPath := filepath.Join("$HOME", config.DefaultConfigDir,
		fmt.Sprintf("%s.%s", config.DefaultConfigName, config.DefaultConfigType))
	rootCmd.PersistentFlags().StringVar(&config.ConfigFile, "config", "", fmt.Sprintf("config file to use (default: '%s')", defaultConfigPath))

	config.BindGlobalFlags(rootCmd.PersistentFlags())
	config.BindGlobalEnv(rootCmd.PersistentFlags())
}
