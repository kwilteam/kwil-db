package utils

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/kwilteam/kwil-db/cmd/internal/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/spf13/cobra"
)

func txQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query-tx TX_ID",
		Short: "Queries a transaction from the blockchain, TX_ID is hex encoded.",
		Long:  longTxQueryDesc,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var resp respTxQuery
			err := common.DialClient(cmd.Context(), common.WithoutPrivateKey, func(ctx context.Context, client *client.Client, conf *config.KwilCliConfig) error {
				txHash, err := hex.DecodeString(args[0])
				if err != nil {
					return fmt.Errorf("error decoding transaction id: %w", err)
				}

				resp.Msg, err = client.TxQuery(ctx, txHash)
				if err != nil {
					return fmt.Errorf("error querying transaction: %w", err)
				}
				return nil
			})

			return display.Print(&resp, err, config.GetOutputFormat())
		},
	}

	return cmd
}

const longTxQueryDesc = `
Queries a transaction from the blockchain. Requires 1 argument: the transaction id.

The transaction id is a hex encoded string.
`
