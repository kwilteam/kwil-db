package fund

import (
	"errors"
	"fmt"
	"kwil/cmd/kwil-cli/cmds/common/display"
	"kwil/cmd/kwil-cli/config"
	escrowTypes "kwil/pkg/chain/contracts/escrow/types"
	"kwil/pkg/client"
	"math/big"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func depositCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "deposit",
		Short: "Deposit funds into the funding pool.",
		Long:  `Deposit funds into the funding pool.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			clt, err := client.New(ctx, config.Config.Node.KwilProviderRpcUrl,
				client.WithChainRpcUrl(config.Config.ClientChain.Provider),
			)
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			// convert arg 0 to big int
			amount, ok := new(big.Int).SetString(args[0], 10)
			if !ok {
				return fmt.Errorf("error converting %s to big int", args[0])
			}

			token, err := clt.TokenContract(ctx)
			if err != nil {
				return fmt.Errorf("failed to get token contract: %w", err)
			}

			fmt.Printf("You will be depositing $%s %s into funding pool %s\n", token.Name(), amount, clt.EscrowContractAddress)
			pr := promptui.Select{
				Label: "Continue?",
				Items: []string{"yes", "no"},
			}

			_, res, err := pr.Run()
			if err != nil {
				return fmt.Errorf("failed to get user input: %w", err)
			}

			if res != "yes" {
				return errors.New("transaction cancelled")
			}

			escrow, err := clt.EscrowContract(ctx)
			if err != nil {
				return fmt.Errorf("failed to get escrow contract: %w", err)
			}

			pk, err := config.GetEcdsaPrivateKey()
			if err != nil {
				return fmt.Errorf("failed to get private key: %w", err)
			}

			txRes, err := escrow.Deposit(ctx, &escrowTypes.DepositParams{
				Amount:    amount,
				Validator: clt.ProviderAddress,
			}, pk)
			if err != nil {
				return fmt.Errorf("failed to send deposit transaction: %w", err)
			}

			display.PrintClientChainResponse(&display.ClientChainResponse{
				Tx:    txRes.TxHash,
				Chain: clt.ChainCode.String(),
			})
			return nil
		},
	}

	return cmd
}
