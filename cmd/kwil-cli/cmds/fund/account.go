package fund

import (
	"context"
	"fmt"
	"kwil/cmd/kwil-cli/cmds/common"
	"kwil/cmd/kwil-cli/config"
	"kwil/pkg/client"

	"github.com/spf13/cobra"
)

func getAccountCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "get-account",
		Short: "Get balance, spent, and nonce information",
		Long: `Gets the balance, spent, and nonce information for a given account address.
If no address is provided, it will use the address of the user's wallet.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialClient(cmd.Context(), common.WithoutServiceConfig, func(ctx context.Context, client *client.Client, conf *config.KwilCliConfig) error {
				// check if config is set
				address, err := getSelectedAddress(cmd, conf)
				if err != nil {
					return fmt.Errorf("error getting selected address: %w", err)
				}

				acc, err := client.GetAccount(ctx, address)
				if err != nil {
					return fmt.Errorf("error getting account config: %w", err)
				}

				fmt.Println("Address: ", acc.Address)
				fmt.Println("Balance: ", acc.Balance)
				fmt.Println("Nonce:   ", acc.Nonce)

				return nil
			})

		},
	}

	cmd.Flags().StringP(addressFlag, "a", "", "Account address to get information for")

	return cmd
}
