package utils

import (
	"context"
	"encoding/hex"
	"fmt"
	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/pkg/abci"
	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/spf13/cobra"
)

func txQueryCmd() *cobra.Command {
	var txidEncoding string // can be either "hex" or "base64"
	cmd := &cobra.Command{
		Use:   "query-tx TX_ID",
		Short: "Queries a transaction from the blockchain, TX_ID is hex encoded.",
		Long:  longTxQueryDesc,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resp := &respTxInfo{}
			err := common.DialClient(cmd.Context(), common.WithoutPrivateKey, func(ctx context.Context, client *client.Client, conf *config.KwilCliConfig) error {
				txHash, err := hex.DecodeString(args[0])
				if err != nil {
					return fmt.Errorf("error decoding transaction id: %w", err)
				}

				res, err := client.TxQuery(ctx, txHash)
				if err != nil {
					return fmt.Errorf("error querying transaction: %w", err)
				}
				resp.Msg = res
				return nil
			})

			msg := display.WrapMsg(resp, err)
			return display.Print(msg, err, config.GetOutputFormat())
		},
	}

	cmd.Flags().StringVarP(&txidEncoding, "encoding", "e", "hex", "the encoding of the transaction id. Can be either 'hex' or 'base64'")
	return cmd
}

const longTxQueryDesc = `
Queries a transaction from the blockchain. Requires 1 argument: the transaction id.

The transaction id can be either a hex or base64 encoded string. The encoding can be specified with the '--encoding' flag (shorthand '-e').
Defaults to hex encoding.
`

type encodeFunc func([]byte) string
type decodeFunc func(string) ([]byte, error)

func printQueryTxRes(res *txpb.TxQueryResponse, encode encodeFunc) {
	fmt.Println("Transaction ID: ", encode(res.Hash))

	status := "failed"
	if res.Height == -1 {
		status = "pending"
	} else if res.TxResult.Code == abci.CodeOk.Uint32() {
		status = "success"
	}

	fmt.Println("Status: ", status)
	fmt.Println("Height: ", res.Height)
	fmt.Println("Outputted Logs: ", res.TxResult.Log)
}
