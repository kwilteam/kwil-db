package account

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/kwilteam/kwil-db/app/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/helpers"
	clientType "github.com/kwilteam/kwil-db/core/client/types"
	"github.com/spf13/cobra"
)

func transferCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer <recipient> <amount>",
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

			return helpers.DialClient(cmd.Context(), cmd, 0, func(ctx context.Context, cl clientType.Client, conf *config.KwilCliConfig) error {
				txHash, err := cl.Transfer(ctx, to, amount, clientType.WithNonce(nonceOverride),
					clientType.WithSyncBroadcast(syncBcast))
				if err != nil {
					return display.PrintErr(cmd, fmt.Errorf("transfer failed: %w", err))
				}
				// If sycnBcast, and we have a txHash (error or not), do a query-tx.
				if len(txHash) != 0 && syncBcast {
					time.Sleep(500 * time.Millisecond) // otherwise it says not found at first
					resp, err := cl.TxQuery(ctx, txHash)
					if err != nil {
						return display.PrintErr(cmd, fmt.Errorf("tx query failed: %w", err))
					}
					return display.PrintCmd(cmd, display.NewTxHashAndExecResponse(resp))
				}
				return display.PrintCmd(cmd, display.RespTxHash(txHash))
			})
		},
	}

	return cmd
}
