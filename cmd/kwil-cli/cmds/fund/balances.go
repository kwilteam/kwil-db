package fund

import (
	"fmt"
	"kwil/cmd/kwil-cli/config"
	"kwil/pkg/client"

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
			ctx := cmd.Context()
			clt, err := client.New(ctx, config.Config.Node.KwilProviderRpcUrl,
				client.WithChainRpcUrl(config.Config.ClientChain.Provider),
			)
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			address, err := getSelectedAddress(cmd)
			if err != nil {
				return fmt.Errorf("error getting selected address: %w", err)
			}

			tokenCtr, err := clt.TokenContract(ctx)
			if err != nil {
				return fmt.Errorf("error getting token contract: %w", err)
			}

			allowance, err := tokenCtr.Allowance(address, clt.EscrowContractAddress)
			if err != nil {
				return fmt.Errorf("error getting allowance: %w", err)
			}

			// get balance
			balance, err := tokenCtr.BalanceOf(address)
			if err != nil {
				return fmt.Errorf("error getting balance: %w", err)
			}

			color.Set(color.Bold)
			cmd.Printf("Pool: %s\n", clt.EscrowContractAddress)
			color.Unset()
			color.Set(color.FgGreen)
			cmd.Printf("Allowance: %s\n", allowance)
			cmd.Printf("Balance: %s\n", balance)
			color.Unset()

			return nil
		},
	}

	cmd.Flags().StringP("account", "a", "", "Account address to get information for")

	return cmd
}
