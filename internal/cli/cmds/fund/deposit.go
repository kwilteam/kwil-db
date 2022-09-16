package fund

import (
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/internal/cli/chain"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func depositCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "deposit",
		Short: "Deposit funds into the funding pool.",
		Long:  `Deposit funds into the funding pool.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := chain.NewClientV(viper.GetViper())
			if err != nil {
				return err
			}

			allowance, err := c.GetAllowance()
			if err != nil {
				return err
			}

			// convert arg 0 to big int
			amount, ok := new(big.Int).SetString(args[0], 10)
			if !ok {
				return fmt.Errorf("error converting %s to big int", args[0])
			}

			// check if allowance >= amount
			if allowance.Cmp(amount) < 0 {
				return fmt.Errorf("not enough tokens to deposit %s (allowance %s)", amount.String(), allowance.String())
			}

			balance, err := c.GetBalance()
			if err != nil {
				return err
			}

			if balance.Cmp(amount) < 0 {
				return fmt.Errorf("not enough tokens to deposit %s (balance %s)", amount.String(), balance.String())
			}

			tokenName, err := c.GetTokenSymbol()
			if err != nil {
				return err
			}

			amtFloat, err := c.ConvertToDecimal(amount)
			if err != nil {
				return err
			}

			_ = tokenName
			_ = amtFloat

			// fmt.Printf("You will be depositing %s %s into funding pool %s\n", amtFloat.String(), tokenName, c.getPoolAddress())
			// res, err := utils.PromptStringInput("Continue? (y/n)")
			// if err != nil {
			// 	return err
			// }

			// if res != "y" {
			// 	fmt.Println("Aborting...")
			// 	return nil
			// }

			// now deposit
			// get the validator address
			vAddr := c.GetValidatorAddress()
			err = c.Deposit(amount, vAddr)
			if err != nil {
				return err
			}
			return nil
		},
	}

	return cmd
}
