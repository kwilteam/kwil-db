package fund

import (
	"context"
	"fmt"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/pkg/client"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func balancesCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "balances",
		Short: "Gets your allowance and deposit balances.",
		Long:  `"balances" returns your allowance and balance for your currently configured funding pool.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialClient(cmd.Context(), common.WithChainClient, func(ctx context.Context, client *client.Client, conf *config.KwilCliConfig) error {
				address, err := getSelectedAddress(cmd, conf)
				if err != nil {
					return fmt.Errorf("error getting selected address: %w", err)
				}

				allowance, err := client.GetApprovedAmount(ctx, address)
				if err != nil {
					return fmt.Errorf("error getting allowance: %w", err)
				}

				// get balance
				balance, err := client.GetOnChainBalance(ctx, address)
				if err != nil {
					return fmt.Errorf("error getting balance: %w", err)
				}

				// get deposited amount
				deposited, err := client.GetDepositedAmount(ctx, address)
				if err != nil {
					return fmt.Errorf("error getting deposited amount: %w", err)
				}

				color.Set(color.Bold)
				fmt.Printf("Pool: %s\n", client.PoolAddress)
				color.Unset()
				color.Set(color.FgGreen)
				fmt.Printf("Allowance: %s\n", allowance)
				fmt.Printf("Balance: %s\n", balance)
				fmt.Printf("Deposit Balance: %s\n", deposited)
				color.Unset()

				return nil
			})
		},
	}

	cmd.Flags().StringP(addressFlag, "a", "", "Account address to get information for")

	return cmd
}
