package fund

import (
	"fmt"
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
			ctx := cmd.Context()
			clt, err := client.New(ctx, config.Config.Node.KwilProviderRpcUrl)
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			// check if config is set
			address, err := getSelectedAddress(cmd)
			if err != nil {
				return fmt.Errorf("error getting selected address: %w", err)
			}

			acc, err := clt.GetAccount(ctx, address)
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

	cmd.Flags().StringP(addressFlag, "a", "", "Account address to get information for")

	return cmd
}
