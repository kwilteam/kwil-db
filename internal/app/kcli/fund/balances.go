package fund

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"kwil/internal/app/kcli/common"
	"kwil/pkg/kwil-client"
)

func balancesCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "balances",
		Short: "Gets your allowance and deposit balances.",
		Long:  `"balances" returns your allowance and balance for your currently configured funding pool.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			clt, err := kwil_client.New(ctx, common.AppConfig)
			if err != nil {
				return err
			}

			allowance, err := clt.Fund.GetAllowance(ctx, clt.Config.Fund.GetAccountAddress(), clt.Config.Fund.PoolAddress)
			if err != nil {
				return fmt.Errorf("error getting allowance: %w", err)
			}

			// get balance
			balance, err := clt.Fund.GetBalance(ctx, clt.Config.Fund.GetAccountAddress())
			if err != nil {
				return fmt.Errorf("error getting deposit balance: %w", err)
			}

			color.Set(color.Bold)
			cmd.Printf("Pool: %s\n", clt.Config.Fund.PoolAddress)
			color.Unset()
			color.Set(color.FgGreen)
			cmd.Printf("Allowance: %s\n", allowance)
			cmd.Printf("Balance: %s\n", balance)
			color.Unset()

			return nil
		},
	}

	return cmd
}
