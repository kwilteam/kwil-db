/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package database

import (
	"fmt"

	"github.com/kwilteam/kwil-db/cli/cmd/utils"
	//"github.com/fatih/color"
	"github.com/kwilteam/kwil-db/cli/cmd/database/queries"
	"github.com/kwilteam/kwil-db/cli/cmd/database/roles"
	"github.com/kwilteam/kwil-db/cli/cmd/database/table"
	"github.com/spf13/cobra"
)

// modifyCmd represents the modify command
var modifyCmd = &cobra.Command{
	Use:   "modify",
	Short: "Modify is used to modify aspects of a database.",
	Long: `With modify, you can make several changes to your database.
	These changes include functions like adding / removing columns,
	defining paramaterized queries, and creating new roles.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Please provide a database name")
			return
		}

		// TODO: check if database exists

		input, err := promptModify(args[0])
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		switch input {
		case "tables":
			table.Table()
		case "roles":
			roles.Roles()
		case "queries":
			queries.Queries()
		}

	},
}

func init() {
	databaseCmd.AddCommand(modifyCmd)
}

func promptModify(s string) (string, error) {
	str := fmt.Sprintf("Please choose what you would like to modify for %s", s)
	return utils.PromptStringArr(str, []string{"Tables", "Roles", "Queries"})
}
