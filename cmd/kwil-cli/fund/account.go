package fund

import (
	"fmt"
	"github.com/spf13/cobra"
	"kwil/cmd/kwil-cli/common"
	"kwil/internal/app/kcli"
)

func getAccountCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "info",
		Short: "Get balance, spent, and nonce information",
		Long:  ``,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			clt, err := kcli.New(ctx, common.AppConfig)
			if err != nil {
				return err
			}

			// check if info is set
			account, err := cmd.Flags().GetString("account")
			if err != nil {
				return fmt.Errorf("error getting account flag: %w", err)
			}

			if account == "" {
				account = clt.Config.Fund.GetAccountAddress()
			}

			acc, err := clt.Client.GetAccount(ctx, account)
			if err != nil {
				return fmt.Errorf("error getting account info: %w", err)
			}
			fmt.Println("Address: ", acc.Address)
			fmt.Println("Balance: ", acc.Balance)
			fmt.Println("Spent:   ", acc.Spent)
			fmt.Println("Nonce:   ", acc.Nonce)

			return nil

		},
	}

	cmd.Flags().StringP("account", "a", "", "Account address to get information for")

	return cmd
}
