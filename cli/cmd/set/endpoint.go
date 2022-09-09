/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package set

import (
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/cli/cmd/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// kwilUrlCmd represents the kwilUrl command
var kwilUrlCmd = &cobra.Command{
	Use:   "endpoint",
	Short: "Set your KwilDB Gateway Endpoint",
	Long: `Set Endpoint allows you to set the URL of your KwilDB Gateway Endpoint.
	It takes in one argument, which is the URL.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := utils.LoadConfig()
		if err != nil {
			fmt.Println(err)
			return
		}

		if len(args) != 1 {
			fmt.Println("set endpoint requires one argument")
			return
		}

		viper.Set("endpoint", args[0])
		viper.WriteConfig()

	},
}

var ethProviderCmd = &cobra.Command{
	Use:   "eth-provider",
	Short: "Set your Ethereum Provider",
	Long: `Set Ethereum Provider allows you to set the URL of your Ethereum Provider.
	It takes in one argument, which is an endpoint.  This endpoint will be used both for REST calls
	and for WebSockets.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := utils.LoadConfig()
		if err != nil {
			fmt.Println(err)
			return
		}

		if len(args) != 1 {
			fmt.Println("set eth-provider requires one argument")
			return
		}

		// check to make sure args[0] doesn't begin with http:// or https:// or ws:// or wss://
		if strings.HasPrefix(args[0], "http://") || strings.HasPrefix(args[0], "https://") || strings.HasPrefix(args[0], "ws://") || strings.HasPrefix(args[0], "wss://") {
			fmt.Println(`The function "set eth-provider" requires an endpoint without a specified protocol.`)
			fmt.Println("For example, instead of \"https://mainnet.infura.io\", use \"mainnet.infura.io\".")
		}

		viper.Set("eth-provider", args[0])
		viper.WriteConfig()

	},
}

func init() {
	setCmd.AddCommand(kwilUrlCmd)
	setCmd.AddCommand(ethProviderCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// kwilUrlCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// kwilUrlCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
