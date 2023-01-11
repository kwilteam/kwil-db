package main

import (
	"fmt"
	"kwil/cmd/kwil-cli/cmds/configure"
	"kwil/cmd/kwil-cli/cmds/connect"
	"kwil/cmd/kwil-cli/cmds/database"
	"kwil/cmd/kwil-cli/cmds/fund"
	"kwil/cmd/kwil-cli/cmds/utils"
	"os"
	"path/filepath"

	"github.com/manifoldco/promptui"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Execute() error {
	cmd := &cobra.Command{
		Use:   "kwil",
		Short: "A brief description of your application",
		Long:  "",
	}

	cobra.OnInitialize(initConfig)

	cmd.AddCommand(
		connect.NewCmdConnect(),
		fund.NewCmdFund(),
		configure.NewCmdConfigure(),
		database.NewCmdDatabase(),
		utils.NewCmdUtils(),
	)

	if err := cmd.Execute(); err != nil {
		if err == promptui.ErrInterrupt {
			err = nil
		}
		return err
	}

	return nil
}

func initConfig() {
	home, err := homedir.Dir()
	if err != nil {
		return
	}
	configFile := filepath.Join(home, ".kwil/config/cli.toml")
	_, err = os.Stat(configFile)
	if err != nil {
		if err := os.MkdirAll(filepath.Dir(configFile), 0755); err != nil {
			fmt.Println(err)
			return
		}

		file, err := os.Create(configFile)
		if err != nil {
			fmt.Println(err)
			return
		}
		file.Close()
	}

	viper.SetConfigFile(configFile)
	_ = viper.ReadInConfig()
}
