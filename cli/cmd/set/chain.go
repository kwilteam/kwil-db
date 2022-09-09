/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package set

import (
	"fmt"

	"github.com/kwilteam/kwil-db/cli/cmd/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// chainCmd represents the chain command
var chainCmd = &cobra.Command{
	Use:   "chain",
	Short: "Allows you to set the chain",
	Long: `"set chain" allows you to set a client chain.
	It gives you a list of supported chains to choose from.`,
	Run: func(cmd *cobra.Command, args []string) {
		res, err := utils.PromptStringArr("Please choose the chain you wish to bridge from", []string{"Ethereum", "Goerli"})
		if err != nil {
			fmt.Println(err)
			return
		}

		// load config
		err = utils.LoadConfig()
		if err != nil {
			fmt.Println(err)
			return
		}

		var cid string

		// now we get chainID
		switch res {
		case "ethereum":
			cid = "1"
		case "goerli":
			cid = "5"
		}

		// set chain
		viper.Set("chain", res)
		viper.Set("chain-id", cid)
		viper.WriteConfig()
	},
}

func init() {
	setCmd.AddCommand(chainCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// chainCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// chainCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
