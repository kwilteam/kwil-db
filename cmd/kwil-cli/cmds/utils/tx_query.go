package utils

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/spf13/cobra"
)

func txQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query-tx <tx_id>",
		Short: "Queries a transaction from the blockchain. Requires 1 argument: the hex encoded transaction id.",
		Long:  `Queries a transaction from the blockchain. Requires 1 argument: the hex encoded transaction id.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialClient(cmd.Context(), cmd, common.WithoutPrivateKey, func(ctx context.Context, client common.Client, conf *config.KwilCliConfig) error {
				txHash, err := hex.DecodeString(args[0])
				if err != nil {
					return display.PrintErr(cmd, fmt.Errorf("error decoding transaction id: %w", err))
				}

				msg, err := client.TxQuery(ctx, txHash)
				if err != nil {
					return display.PrintErr(cmd, fmt.Errorf("error querying transaction: %w", err))
				}
				return display.PrintCmd(cmd, &respTxQuery{Msg: msg})
			})

		},
	}

	return cmd
}
