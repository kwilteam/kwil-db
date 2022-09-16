package fund

import (
	"fmt"

	"github.com/kwilteam/kwil-db/internal/cli/chain"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func balancesCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "balances",
		Short: "Gets your allowance and deposit balances.",
		Long:  `"balances" returns your allowance and balance for your currently configured funding pool.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := chain.NewClientV(viper.GetViper())
			if err != nil {
				return err
			}

			allowance, err := c.GetAllowance()
			if err != nil {
				return fmt.Errorf("error getting allowance: %w", err)
			}

			// convert allowance to float
			af, err := c.ConvertToDecimal(allowance)
			if err != nil {
				return fmt.Errorf("error converting allowance to decimal: %w", err)
			}

			sym, err := c.GetTokenSymbol()
			if err != nil {
				return fmt.Errorf("error getting symbol: %w", err)
			}

			// get balance
			balance, err := c.GetDepositBalance()
			if err != nil {
				return fmt.Errorf("error getting deposit balance: %w", err)
			}

			// convert balance to float
			bf, err := c.ConvertToDecimal(balance)
			if err != nil {
				return fmt.Errorf("error converting balance to decimal: %w", err)
			}

			cmd.Printf("Pool: %s\n", c.GetPoolAddress())
			cmd.Printf("Allowance: %s (%s %s)\n", allowance, af.String(), sym)
			cmd.Printf("Balance: %s (%s %s)\n", balance, bf.String(), sym)

			return nil
		},
	}

	return cmd
}
