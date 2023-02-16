package fund

import (
	"errors"
	"fmt"
	"kwil/cmd/kwil-cli/cmds/common/display"
	"kwil/cmd/kwil-cli/conf"
	"kwil/pkg/client"
	"math/big"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func approveCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "approve",
		Short: "Approves the funding pool to spend your tokens",
		Long:  `Approves the funding pool to spend your tokens.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			clt, err := client.New(ctx, conf.Config.Node.KwilProviderRpcUrl,
				client.WithChainRpcUrl(conf.Config.ClientChain.Provider),
			)
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			amount, ok := new(big.Int).SetString(args[0], 10)
			if !ok {
				return fmt.Errorf("could not convert %s to int", args[0])
			}

			cmd.Printf("You will have a new amount approved of: %s\n", amount.String())

			// ask one more time to confirm the transaction
			pr := promptui.Select{
				Label: "Continue?",
				Items: []string{"yes", "no"},
			}

			_, res, err := pr.Run()
			if err != nil {
				return err
			}

			if res != "yes" {
				return errors.New("transaction cancelled")
			}

			tokenCtr, err := clt.TokenContract(ctx)
			if err != nil {
				return err
			}

			pk, err := conf.GetEcdsaPrivateKey()
			if err != nil {
				return err
			}

			response, err := tokenCtr.Approve(ctx, clt.EscrowContractAddress, amount, pk)
			if err != nil {
				return err
			}

			display.PrintClientChainResponse(&display.ClientChainResponse{
				Tx:    response.TxHash,
				Chain: clt.ChainCode.String(),
			})

			return nil
		},
	}

	return cmd
}
