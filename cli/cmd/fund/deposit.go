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

// depositCmd represents the deposit command
var depositCmd = &cobra.Command{
	Use:   "deposit",
	Short: "Deposit funds into the funding pool.",
	Long:  `Deposit funds into the funding pool.`,
	Run: func(cmd *cobra.Command, args []string) {
		// check to make sure ther is one arg
		if len(args) != 1 {
			fmt.Println("fund deposit requires one argument")
			return
		}

		// load config
		err := utils.LoadConfig()
		if err != nil {
			fmt.Println(err)
			return
		}

		// create new client
		c, err := newChainClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		// check that the amount they wish to deposit is greater than allowance

		allowance, err := c.getAllowance()
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

		// check if allowance >= amount
		if allowance.Cmp(amount) < 0 {
			fmt.Printf("Not enough allowance to fund.  You have %s tokens approved, but you are trying to fund %s tokens.\n", allowance, amount)
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

		// get token name
		tokenName, err := c.getTokenSymbol()
		if err != nil {
			fmt.Println(err)
			return
		}

		// convert to float
		amtFloat, err := c.convertToDecimal(amount)
		if err != nil {
			fmt.Println(err)
			return
		}

		// ask for confirmation:
		fmt.Printf("You will be depositing %s %s into funding pool %s\n", amtFloat.String(), tokenName, c.getPoolAddress())
		res, err := utils.PromptStringInput("Continue? (y/n)")
		if err != nil {
			fmt.Println(err)
			return
		}

		if res != "y" {
			fmt.Println("Aborting...")
			return
		}

		// now deposit
		// get the validator address
		vAddr := c.getValidatorAddress()
		err = c.deposit(amount, vAddr)
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

func init() {
	fundCmd.AddCommand(depositCmd)
}
