package main

import (
	"fmt"
	"kwil/cmd/kwil-cli/common"
	"kwil/cmd/kwil-cli/configure"
	"kwil/cmd/kwil-cli/database"
	"kwil/cmd/kwil-cli/fund"
	"kwil/cmd/kwil-cli/utils"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func init() {
	cobra.OnInitialize(common.LoadConfig)
}

func Execute() error {

	var cmd = &cobra.Command{
		Use:   "kwil",
		Short: "A kwil command line interface",
		Long:  "",
	}

	cobra.OnInitialize(common.LoadConfig)
	common.LoadConfig()

	cmd.AddCommand(
		fund.NewCmdFund(),
		configure.NewCmdConfigure(),
		database.NewCmdDatabase(),
		utils.NewCmdUtils(),
	)

	common.BindGlobalFlags(cmd.PersistentFlags())
	common.BindGlobalEnv(cmd.PersistentFlags())

	if err := cmd.Execute(); err != nil {
		if err == promptui.ErrInterrupt {
			err = nil
		}
		return err
	}

	return nil
}

func main() {
	if err := Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
