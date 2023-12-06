package account

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/core/client"

	"github.com/spf13/cobra"
)

func transferCmd() *cobra.Command {
	// var recipient, amt string

	cmd := &cobra.Command{
		Use:   "transfer",
		Short: "Transfer value to an account",
		Long:  `Transfers value to an account.`,
		Args:  cobra.ExactArgs(2), // recipient, amt
		RunE: func(cmd *cobra.Command, args []string) error {
			recipient, amt := args[0], args[1]
			to, err := hex.DecodeString(recipient) // identifier bytes
			if err != nil {
				return display.PrintErr(cmd, err)
			}
			amount, ok := big.NewInt(0).SetString(amt, 10)
			if !ok {
				return display.PrintErr(cmd, errors.New("invalid decimal amount"))
			}

			var txHash []byte
			err = common.DialClient(cmd.Context(), cmd, 0, func(ctx context.Context, cl common.Client, conf *config.KwilCliConfig) error {
				txHash, err = cl.Transfer(ctx, to, amount, client.WithNonce(nonceOverride))
				if err != nil {
					return display.PrintErr(cmd, fmt.Errorf("transfer failed: %w", err))
				}

				return nil
			})

			return display.PrintCmd(cmd, display.RespTxHash(txHash))
		},
	}

	// const (
	// 	toFlagName  = "to"
	// 	amtFlagName = "amount"
	// )
	// cmd.Flags().StringVar(&recipient, toFlagName, "", "the recipient (required)")
	// cmd.Flags().StringVar(&amt, amtFlagName, "", "the recipient (required)")

	// cmd.MarkFlagRequired(toFlagName)

	return cmd
}
