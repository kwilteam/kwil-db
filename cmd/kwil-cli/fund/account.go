package fund

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"kwil/cmd/kwil-cli/common"
	"kwil/internal/app/kcli"
	"kwil/pkg/fund"
)

func getAccountCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "account",
		Short: "Gets account balance, spent, and nonce information",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialGrpc(cmd.Context(), func(ctx context.Context, cc *grpc.ClientConn) error {
				conf, err := fund.NewConfig()
				if err != nil {
					return fmt.Errorf("error getting client config: %w", err)
				}

				client, err := kcli.New(cc, conf)
				if err != nil {
					return fmt.Errorf("error creating client: %w", err)
				}

				// check if account is set
				account, err := cmd.Flags().GetString("account")
				if err != nil {
					return fmt.Errorf("error getting account flag: %w", err)
				}

				if account == "" {
					account = client.Fund.GetConfig().GetAccountAddress()
				}

				acc, err := client.Accounts.GetAccount(ctx, account)
				if err != nil {
					return fmt.Errorf("error getting account: %w", err)
				}

				fmt.Println("Address: ", acc.Address)
				fmt.Println("Balance: ", acc.Balance)
				fmt.Println("Spent:   ", acc.Spent)
				fmt.Println("Nonce:   ", acc.Nonce)

				return nil
			},
			)
		},
	}

	cmd.Flags().StringP("account", "a", "", "Account address to get information for")

	return cmd
}
