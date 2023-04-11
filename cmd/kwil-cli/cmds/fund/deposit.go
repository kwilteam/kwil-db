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

func depositCmd() *cobra.Command {
	var opts struct {
		assumeYes bool
	}

	var cmd = &cobra.Command{
		Use:   "deposit AMOUNT",
		Short: "Deposit funds into the funding pool.",
		Long:  `Deposit funds into the funding pool.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialClient(cmd.Context(), common.WithChainClient, func(ctx context.Context, client *client.Client, config *config.KwilCliConfig) error {
				// convert arg 0 to big int
				amount, ok := new(big.Int).SetString(args[0], 10)
				if !ok {
					return fmt.Errorf("error converting %s to big int", args[0])
				}

				if !opts.assumeYes {
					fmt.Printf("You will be depositing $%s %s into funding pool %s\n", client.TokenSymbol, amount, client.PoolAddress)
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
				}

				txHash, err := client.Deposit(ctx, amount)
				if err != nil {
					return fmt.Errorf("error depositing funds: %w", err)
				}

				display.PrintClientChainResponse(&display.ClientChainResponse{
					Tx:    txHash,
					Chain: client.ChainCode.String(),
				})
				return nil
			})
		},
	}

	flags := cmd.Flags()
	flags.BoolVarP(&opts.assumeYes, "yes", "y", false, "Automatic yes to prompts.")

	return cmd
}
