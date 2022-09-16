package fund

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/internal/cli/chain"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func withdrawCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "withdraw",
		Short: "Withdraws funds from the funding pool",
		Long:  `"withdraw" withdraws funds from the funding pool.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := chain.NewClientV(viper.GetViper())
			if err != nil {
				return err
			}

			// get balance
			balance, err := c.GetDepositBalance()
			if err != nil {
				return fmt.Errorf("error getting deposit balance: %w", err)
			}

			amount, ok := new(big.Int).SetString(args[0], 10)
			if !ok {
				return errors.New("could not convert amount to big int")
			}

			if balance.Cmp(amount) < 0 {
				return fmt.Errorf("insufficient funds: %s of %s", amount, balance)
			}

			// now withdraw
			err = c.Withdraw(amount)
			if err != nil {
				return fmt.Errorf("error withdrawing: %w", err)
			}
			return nil
		},
	}

	return cmd
}
