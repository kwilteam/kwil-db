/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package fund

import (
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/cli/cmd/utils"
	"github.com/spf13/cobra"
)

// approveCmd represents the approve command
var approveCmd = &cobra.Command{
	Use:   "approve",
	Short: "Approves the funding pool to spend your tokens",
	Long:  `Approves the funding pool to spend your tokens.`,
	Run: func(cmd *cobra.Command, args []string) {
		// check to make sure there is one arg
		if len(args) != 1 {
			fmt.Println("fund approve requires one argument")
			return
		}

		// load config
		err := utils.LoadConfig()
		if err != nil {
			fmt.Println(err)
			return
		}

		// convert arg 0 to big int
		amount, ok := new(big.Int).SetString(args[0], 10)
		if !ok {
			fmt.Println("could not convert amount to big int")
			return
		}

		// make chain client
		c, err := newChainClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		// now get balance
		balance, err := c.getBalance()
		if err != nil {
			fmt.Println(err)
			return
		}

		// check if balance >= amount
		if balance.Cmp(amount) < 0 {
			fmt.Printf("Not enough tokens to fund.  You have %s tokens, but you are trying to fund %s tokens.\n", balance, amount)
			return
		}

		fmt.Printf("You will have a new amount approved of: %s\nYour token balance: %s\n", amount.String(), balance.String())

		// ask one more time to confirm the transaction
		res, err := utils.PromptStringInput("Continue? (y/n)")
		if err != nil {
			fmt.Println(err)
			return
		}
		if res != "y" {
			fmt.Println("transaction cancelled")
			return
		}

		// approve
		err = c.approve(amount)
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

func init() {
	fundCmd.AddCommand(approveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// approveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// approveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
