package fund

import (
	"context"
	"fmt"
	"kwil/cmd/kwil-cli/common"
	grpc_client "kwil/kwil/client/grpc-client"
	"kwil/x/fund"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

func balancesCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "balances",
		Short: "Gets your allowance and deposit balances.",
		Long:  `"balances" returns your allowance and balance for your currently configured funding pool.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialGrpc(cmd.Context(), func(ctx context.Context, cc *grpc.ClientConn) error {
				// @yaiba TODO: no need to dial grpc here, just use the chain client
				conf, err := fund.NewConfig()
				if err != nil {
					return fmt.Errorf("error getting client config: %w", err)
				}

				client, err := grpc_client.NewClient(cc, conf)
				if err != nil {
					return err
				}

				allowance, err := client.Chain.GetAllowance(ctx, client.Chain.GetConfig().GetAccount(), client.Chain.GetConfig().PoolAddress)
				if err != nil {
					return fmt.Errorf("error getting allowance: %w", err)
				}

				// get balance
				balance, err := client.Chain.GetBalance(ctx, client.Chain.GetConfig().GetAccount())
				if err != nil {
					return fmt.Errorf("error getting deposit balance: %w", err)
				}

				color.Set(color.Bold)
				cmd.Printf("Pool: %s\n", client.Chain.GetConfig().PoolAddress)
				color.Unset()
				color.Set(color.FgGreen)
				cmd.Printf("Allowance: %s\n", allowance)
				cmd.Printf("Balance: %s\n", balance)
				color.Unset()

				return nil

			})

		},
	}

	return cmd
}
