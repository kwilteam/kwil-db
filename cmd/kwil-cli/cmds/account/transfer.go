package account

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/kwilteam/kwil-db/app/shared/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/client"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	clientType "github.com/kwilteam/kwil-db/core/client/types"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/spf13/cobra"
)

func transferCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer <recipientID> <recipientKeyType> <amount>",
		Short: "Transfer value to an account",
		Long:  `Transfers value to an account.`,
		Args:  cobra.ExactArgs(3), // recipient, keytype, amt
		RunE: func(cmd *cobra.Command, args []string) error {
			recipient, typeStr, amt := args[0], args[1], args[2]
			amount, ok := big.NewInt(0).SetString(amt, 10)
			if !ok {
				return display.PrintErr(cmd, errors.New("invalid decimal amount"))
			}

			// Recognize 0x prefix to permit ethereum address format rather
			// than compact ID hex bytes.
			recipient = strings.TrimPrefix(recipient, "0x")
			id, err := hex.DecodeString(recipient)
			if err != nil {
				return display.PrintErr(cmd, fmt.Errorf("failed to decode account ID: %w", err))
			}

			keyType := crypto.KeyType(typeStr)
			// NOTE: could validate on client side first if built with extensions:
			//   keyType, err := crypto.ParseKeyType(typeStr)
			// Otherwise we leave it to the nodes to decide if it is supported.

			to := &types.AccountID{
				Identifier: id,
				KeyType:    keyType,
			}

			return client.DialClient(cmd.Context(), cmd, 0, func(ctx context.Context, cl clientType.Client, conf *config.KwilCliConfig) error {
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
