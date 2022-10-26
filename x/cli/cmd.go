package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"kwil/x/cli/cmds/configure"
	"kwil/x/cli/cmds/connect"
	"kwil/x/cli/cmds/database"
	"kwil/x/cli/cmds/fund"
	"kwil/x/cli/cmds/role"
	"kwil/x/cli/cmds/schema"
	"kwil/x/cli/cmds/table"

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
		table.NewCmdTable(),
		role.NewCmdRole(),
		schema.NewCmdSchema(),
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
