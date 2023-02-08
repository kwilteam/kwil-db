package fund

import (
	"fmt"
	"github.com/spf13/cobra"
	"kwil/internal/app/kcli/config"
	"kwil/pkg/kclient"
)

func getAccountCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "config",
		Short: "Get balance, spent, and nonce information",
		Long:  ``,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			clt, err := kclient.New(ctx, config.AppConfig)
			if err != nil {
				return err
			}

			// check if config is set
			account, err := cmd.Flags().GetString("account")
			if err != nil {
				return fmt.Errorf("error getting account flag: %w", err)
			}

			if account == "" {
				account = clt.Config.Fund.GetAccountAddress()
			}

			acc, err := clt.Kwil.GetAccount(ctx, account)
			if err != nil {
				return fmt.Errorf("error getting account config: %w", err)
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
