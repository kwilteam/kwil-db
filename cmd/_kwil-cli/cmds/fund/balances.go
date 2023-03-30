package fund

import (
	"fmt"
	ec "github.com/ethereum/go-ethereum/crypto"
	"kwil/cmd/kwil-cli/config"
	escrowTypes "kwil/pkg/chain/contracts/escrow/types"
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
				client.WithChainRpcUrl(config.Config.ClientChain.ProviderRpcUrl),
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

			// get deposit balance
			pk, err := config.GetEcdsaPrivateKey()
			if err != nil {
				return fmt.Errorf("failed to get private key: %w", err)
			}
			escrowCtr, err := clt.EscrowContract(ctx)
			depositBalance, err := escrowCtr.Balance(ctx, &escrowTypes.DepositBalanceParams{
				Validator: clt.ProviderAddress,
				Address:   ec.PubkeyToAddress(pk.PublicKey).Hex()})
			if err != nil {
				return fmt.Errorf("error getting deposit balance: %w", err)
			}

			color.Set(color.Bold)
			fmt.Printf("Pool: %s\n", clt.EscrowContractAddress)
			color.Unset()
			color.Set(color.FgGreen)
			fmt.Printf("Allowance: %s\n", allowance)
			fmt.Printf("Balance: %s\n", balance)
			fmt.Printf("Deposit Balance: %s\n", depositBalance.Balance)
			color.Unset()

			return nil
		},
	}

	cmd.Flags().StringP(addressFlag, "a", "", "Account address to get information for")

	return cmd
}
