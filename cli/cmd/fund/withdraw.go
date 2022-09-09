/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package fund

import (
	"fmt"

	"github.com/kwilteam/kwil-db/cli/cmd/utils"
	"github.com/spf13/cobra"
	"math/big"
)

// withdrawCmd represents the withdraw command
var withdrawCmd = &cobra.Command{
	Use:   "withdraw",
	Short: "Withdraws funds from the funding pool",
	Long:  `"withdraw" withdraws funds from the funding pool.`,
	Run: func(cmd *cobra.Command, args []string) {
		// check to make sure there is one arg
		if len(args) != 1 {
			fmt.Println("fund withdraw requires one argument")
			return
		}

		// load config
		err := utils.LoadConfig()
		if err != nil {
			fmt.Println("error loading config:", err)
			return
		}

		c, err := newChainClient()
		if err != nil {
			fmt.Println("error creating chain client:", err)
			return
		}

		// get balance
		balance, err := c.getDepositBalance()
		if err != nil {
			fmt.Println("error getting deposit balance:", err)
			return
		}

		// check if arg 0 is less than or equal to balance
		amount, ok := new(big.Int).SetString(args[0], 10)
		if !ok {
			fmt.Println("could not convert amount to big int")
			return
		}

		if balance.Cmp(amount) < 0 {
			fmt.Printf("Not enough balance to withdraw.  You have %s tokens in the pool, but you are trying to withdraw %s tokens.\n", balance, amount)
			return
		}

		// now withdraw
		err = c.withdraw(amount)
		if err != nil {
			fmt.Println("error withdrawing:", err)
			return
		}
	},
}

func init() {
	fundCmd.AddCommand(withdrawCmd)
}
