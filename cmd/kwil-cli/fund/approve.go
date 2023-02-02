package fund

import (
	"errors"
	"fmt"
	"kwil/cmd/kwil-cli/common"
	"kwil/cmd/kwil-cli/common/display"
	"kwil/internal/app/kcli"
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

			clt, err := kcli.New(ctx, common.AppConfig)
			if err != nil {
				return err
			}

			response, err := clt.Fund.ApproveToken(ctx, clt.Config.Fund.PoolAddress, amount)
			if err != nil {
				return err
			}

			display.PrintClientChainResponse(&display.ClientChainResponse{
				Tx:    response.TxHash,
				Chain: string(clt.Config.Fund.ChainCode),
			})

			return nil
		},
	}

	return cmd
}
