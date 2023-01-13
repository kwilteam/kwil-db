package fund

import (
	"context"
	"fmt"
	"kwil/cmd/kwil-cli/common"
	"kwil/kwil/client/grpc-client"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func balancesCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "balances",
		Short: "Gets your allowance and deposit balances.",
		Long:  `"balances" returns your allowance and balance for your currently configured funding pool.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialGrpc(cmd.Context(), viper.GetViper(), func(ctx context.Context, cc *grpc.ClientConn) error {
				client, err := grpc_client.NewClient(cc, viper.GetViper())
				if err != nil {
					return err
				}

				allowance, err := client.Token.Allowance(client.Config.Address, client.Config.PoolAddress)
				if err != nil {
					return fmt.Errorf("error getting allowance: %w", err)
				}

				sym := client.Token.Symbol()

				// get balance
				balance, err := client.Token.BalanceOf(client.Config.Address)
				if err != nil {
					return fmt.Errorf("error getting deposit balance: %w", err)
				}

				color.Set(color.Bold)
				cmd.Printf("Pool: %s\n", client.Config.PoolAddress)
				color.Unset()
				color.Set(color.FgGreen)
				cmd.Printf("Allowance: %s %s\n", allowance, sym)
				cmd.Printf("Balance: %s %s\n", balance, sym)
				color.Unset()

				return nil

			})

		},
	}

	return cmd
}
