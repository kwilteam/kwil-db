/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package database

import (
	"fmt"

	"github.com/kwilteam/kwil-db/cli/cmd/utils"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create is used for creating a new database.",
	Long: `Create is used for creating a new database that will be stored
	under your account.  It takes in one argument, which is the name of the database.`,
	Run: func(cmd *cobra.Command, args []string) {
		dbName, err := utils.PromptStringInput("Name")
		if err != nil {
			fmt.Println(err)
			return
		}

		bbName, err := utils.PromptStringInput("Bucket Name")
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(dbName)
		fmt.Println(bbName)

	},
}

func init() {
	databaseCmd.AddCommand(createCmd)
}
