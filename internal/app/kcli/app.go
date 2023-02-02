package kcli

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	common "kwil/internal/app/kcli/common"
	"kwil/internal/app/kcli/database"
	"kwil/internal/app/kcli/fund"
	initCli "kwil/internal/app/kcli/init"
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

	rootCmd.SilenceUsage = true
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(common.LoadConfig)

	defaultConfigPath := filepath.Join("$HOME", common.DefaultConfigDir,
		fmt.Sprintf("%s.%s", common.DefaultConfigName, common.DefaultConfigType))
	rootCmd.PersistentFlags().StringVar(&common.ConfigFile, "config", "", fmt.Sprintf("config file (default is %s)", defaultConfigPath))

	bindGlobalFlags(rootCmd.PersistentFlags())
	bindGlobalEnv(rootCmd.PersistentFlags())
}

// bindGlobalFlags binds the global flags to the command.
func bindGlobalFlags(fs *pflag.FlagSet) {
	fs.String("node.endpoint", "", "the endpoint of the Kwil node")
	fs.String("fund.chain_code", "", "the chain code of the funding pool chain")
	fs.String("fund.token_address", "", "the address of the funding pool token")
	fs.String("fund.pool_address", "", "the address of the funding pool")
	fs.String("fund.validator_address", "", "the address of the funding pool validator")
	fs.String("fund.provider", "", "the provider of the funding pool")
	fs.Int64("fund.reconnect_interval", 0, "the reconnect interval of the funding pool")
	fs.Int64("fund.block_confirmation", 0, "the block confirmation of the funding pool")
}

// BindGlobalEnv binds the global flags to the environment variables.
func bindGlobalEnv(fs *pflag.FlagSet) {
	// node.endpoint maps to PREFIX_NODE_ENDPOINT
	viper.SetEnvPrefix(common.EnvPrefix)

	viper.BindEnv("node.endpoint")
	viper.BindPFlag("node.endpoint", fs.Lookup("node.endpoint")) //flag override env

	viper.BindEnv("fund.chain_code")
	viper.BindPFlag("fund.chain_code", fs.Lookup("fund.chain_code"))

	viper.BindEnv("fund.token_address")
	viper.BindPFlag("fund.token_address", fs.Lookup("fund.token_address"))

	viper.BindEnv("fund.pool_address")
	viper.BindPFlag("fund.pool_address", fs.Lookup("fund.pool_address"))

	viper.BindEnv("fund.validator_address")
	viper.BindPFlag("fund.validator_address", fs.Lookup("fund.validator_address"))

	viper.BindEnv("fund.provider")
	viper.BindPFlag("fund.provider", fs.Lookup("fund.provider"))

	viper.BindEnv("fund.reconnect_interval")
	viper.BindPFlag("fund.reconnect_interval", fs.Lookup("fund.reconnect_interval"))

	viper.BindEnv("fund.block_confirmation")
	viper.BindPFlag("fund.block_confirmation", fs.Lookup("fund.block_confirmation"))
}
