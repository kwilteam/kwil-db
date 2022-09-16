package fund

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/internal/cli/chain"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func approveCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "approve",
		Short: "Approves the funding pool to spend your tokens",
		Long:  `Approves the funding pool to spend your tokens.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			amount, ok := new(big.Int).SetString(args[0], 10)
			if !ok {
				return fmt.Errorf("could not convert %s to int", args[0])
			}

			c, err := chain.NewClientV(viper.GetViper())
			if err != nil {
				return err
			}

			// now get balance
			balance, err := c.GetBalance()
			if err != nil {
				return err
			}

			// check if balance >= amount
			if balance.Cmp(amount) < 0 {
				return fmt.Errorf("not enough tokens to fund %s (balance %s)", amount.String(), balance.String())
			}

			cmd.Printf("You will have a new amount approved of: %s\nYour token balance: %s\n", amount.String(), balance.String())

			// ask one more time to confirm the transaction
			// res, err := utils.PromptStringInput("Continue? (y/n)")
			// if err != nil {
			// 	return err
			// }
			res := "n"

			if res != "y" {
				return errors.New("transaction cancelled")
			}

			// approve
			err = c.Approve(amount)
			if err != nil {
				return err
			}
			return nil
		},
	}

	return cmd
}
