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

// apiKeyCmd represents the apiKey command
var apiKeyCmd = &cobra.Command{
	Use:   "api-key",
	Short: "Sets an api key",
	Long:  `Set api-key allows you to set an api key.  It takes in one argument, which is the api key.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := utils.LoadConfig()
		if err != nil {
			fmt.Println(err)
			return
		}

		if len(args) != 1 {
			fmt.Println("set api-key requires one argument")
			return
		}

		viper.Set("api-key", args[0])
		viper.WriteConfig()
	},
}

func init() {
	setCmd.AddCommand(apiKeyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// apiKeyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// apiKeyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
