package root

import (
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"kwil/internal/app/kcli"
	"os"
	"path/filepath"
)

const (
	// DefaultConfigPath is the default path to the config file
	DefaultConfigPath = "$HOME/.kwil/config/cli.toml"
	EnvPrefix         = "KWIL"
)

var cfgFile string
var cfg kcli.Config

var rootCmd = &cobra.Command{
	Use:   "kwil",
	Short: "Kwil command line interface",
	Long:  "Kwil cli allows you to interact with the Kwil",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(os.Stdout, "Kwil cli allows you to interact with the Kwil")
	},
}

func Execute() error {
	//rootCmd.AddCommand(
	//	fund.NewCmdFund(),
	//	configure.NewCmdConfigure(),
	//	database.NewCmdDatabase(),
	//	utils.NewCmdUtils(),
	//	init.NewCmdInit(),
	//)
	//
	//common.BindGlobalFlags(rootCmd.PersistentFlags())
	//common.BindGlobalEnv(rootCmd.PersistentFlags())

	if err := rootCmd.Execute(); err != nil {
		if err == promptui.ErrInterrupt {
			err = nil
		}
		return err
	}

	return nil
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("config file (default is %s)", DefaultConfigPath))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// search config in home/.kwil/config directory with name "cli" (without extension)
		viper.AddConfigPath(filepath.Join(home, ".kwil", "config"))
		viper.SetConfigName("cli")
		viper.SetConfigType("toml")
	}

	viper.SetEnvPrefix(EnvPrefix)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stdout, "Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Fprintln(os.Stderr, "Error loading config file:", err)
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		fmt.Fprintln(os.Stderr, "Error unmarshaling config file:", err)
	}

	viper.WriteConfig()
}
