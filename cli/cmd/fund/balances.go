/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package fund

import (
	"fmt"

	"github.com/kwilteam/kwil-db/cli/cmd/utils"
	"github.com/spf13/cobra"
)

// allowanceCmd represents the allowance command
var allowanceCmd = &cobra.Command{
	Use:   "balances",
	Short: "Gets your allowance and deposit balances.",
	Long: `"balances" returns your allowance and balance for your currently configured
	funding pool.`,
	Run: func(cmd *cobra.Command, args []string) {
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

		allowance, err := c.getAllowance()
		if err != nil {
			fmt.Println("error getting allowance:", err)
			return
		}

		// convert allowance to float
		af, err := c.convertToDecimal(allowance)
		if err != nil {
			fmt.Println("error converting allowance to decimal:", err)
			return
		}

		sym, err := c.getTokenSymbol()
		if err != nil {
			fmt.Println("error getting symbol:", err)
			return
		}

		// get balance
		balance, err := c.getDepositBalance()
		if err != nil {
			fmt.Println("error getting deposit balance:", err)
			return
		}

		// convert balance to float
		bf, err := c.convertToDecimal(balance)
		if err != nil {
			fmt.Println("error converting balance to decimal:", err)
			return
		}

		fmt.Printf("Pool: %s\n", c.poolAddr.Hex())
		fmt.Printf("Allowance: %s (%s %s)\n", allowance, af.String(), sym)
		fmt.Printf("Balance: %s (%s %s)\n", balance, bf.String(), sym)
	},
}

func init() {
	fundCmd.AddCommand(allowanceCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// allowanceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// allowanceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
