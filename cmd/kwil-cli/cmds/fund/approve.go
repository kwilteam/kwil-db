package fund

import (
	"context"
	"errors"
	"fmt"
	"kwil/cmd/kwil-cli/cmds/common"
	"kwil/cmd/kwil-cli/cmds/common/display"
	"kwil/cmd/kwil-cli/config"
	"kwil/pkg/client"
	"math/big"

	"github.com/manifoldco/promptui"

	"github.com/spf13/cobra"
)

func approveCmd() *cobra.Command {
	var opts struct {
		assumeYes bool
	}

	var cmd = &cobra.Command{
		Use:   "approve AMOUNT",
		Short: "Approves the funding pool to spend your tokens",
		Long:  `Approves the funding pool to spend your tokens.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialClient(cmd.Context(), common.WithChainClient, func(ctx context.Context, client *client.Client, config *config.KwilCliConfig) error {
				amount, ok := new(big.Int).SetString(args[0], 10)
				if !ok {
					return fmt.Errorf("could not convert %s to int", args[0])
				}

				fmt.Printf("You will have a new amount approved of: %s\n", amount.String())

				if !opts.assumeYes {
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
				}

				txHash, err := client.ApproveDeposit(ctx, amount)
				if err != nil {
					return fmt.Errorf("error approving deposit: %w", err)
				}

				display.PrintClientChainResponse(&display.ClientChainResponse{
					Tx:    txHash,
					Chain: client.ChainCode.String(),
				})

				return nil
			})
		},
	}

	cmd.Flags().BoolVarP(&opts.assumeYes, "yes", "y", false, "Automatic yes to prompts.")

	return cmd
}
