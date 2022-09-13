/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package set

import (
	"fmt"

	"github.com/kwilteam/kwil-db/internal/cli/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// privateKeyCmd represents the privateKey command
var privateKeyCmd = &cobra.Command{
	Use:   "private-key",
	Short: "Sets your private key",
	Long:  `"set private-key" allows you to set your private key.  It takes in one argument, which is the private key.`,
	Run: func(cmd *cobra.Command, args []string) {
		// check to make sure ther is one arg
		if len(args) != 1 {
			fmt.Println("set private-key requires one argument")
			return
		}

		// load config
		err := utils.LoadConfig()
		if err != nil {
			fmt.Println(err)
			return
		}

		// set private key
		viper.Set("private-key", args[0])
		viper.WriteConfig()

		// check to make sure the arg is a valid private key
	},
}

func init() {
	setCmd.AddCommand(privateKeyCmd)
}
