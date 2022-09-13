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

// fundingPoolCmd represents the fundingPool command
var fundingPoolCmd = &cobra.Command{
	Use:   "funding-pool",
	Short: "Sets the funding pool address",
	Long: `funding-pool allows you to set the funding pool address.
	It takes in one argument, which is the funding pool address.`,
	Run: func(cmd *cobra.Command, args []string) {
		// check to make sure ther is one arg
		if len(args) != 1 {
			fmt.Println("set funding-pool requires one argument")
			return
		}

		// load config
		err := utils.LoadConfig()
		if err != nil {
			fmt.Println(err)
			return
		}

		// set funding pool
		viper.Set("funding-pool", args[0])
		if err = viper.WriteConfig(); err != nil {
			fmt.Println(err)
			return
		}
	},
}

func init() {
	setCmd.AddCommand(fundingPoolCmd)
}
