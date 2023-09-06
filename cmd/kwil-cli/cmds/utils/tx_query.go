package utils

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/pkg/abci"
	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/spf13/cobra"
)

func txQueryCmd() *cobra.Command {
	var txidEncoding string // can be either "hex" or "base64"
	cmd := &cobra.Command{
		Use:   "query-tx TX_ID",
		Short: "Queries a transaction from the blockchain",
		Long:  longTxQueryDesc,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialClient(cmd.Context(), common.WithoutPrivateKey, func(ctx context.Context, client *client.Client, conf *config.KwilCliConfig) error {
				var encode encodeFunc
				var decode decodeFunc

				txString := args[0]

				switch txidEncoding {
				case "hex", "":
					if len(txString) < 2 {
						return fmt.Errorf("invalid transaction id: %s", txString)
					}
					if txString[1] == 'x' {
						txString = txString[2:]
					}

					encode = hex.EncodeToString
					decode = hex.DecodeString
				case "base64":
					encode = base64.StdEncoding.EncodeToString
					decode = base64.StdEncoding.DecodeString
				default:
					return fmt.Errorf("invalid encoding: %s", txidEncoding)
				}

				encodedTxid, err := decode(txString)
				if err != nil {
					return fmt.Errorf("error decoding transaction id: %w", err)
				}

				res, err := client.TxQuery(ctx, encodedTxid)
				if err != nil {
					return fmt.Errorf("error querying transaction: %w", err)
				}

				printQueryTxRes(res, encode)

				return nil
			})
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
	if res.TxResult.Code == abci.CodeOk.Uint32() {
		status = "success"
	}

	fmt.Println("Status: ", status)
	fmt.Println("Data: ", hex.EncodeToString(res.TxResult.Data))
	fmt.Println("Outputted Logs: ", res.TxResult.Log)
}

/*
	d.logger.Info("tx info", zap.Uint64("height", resp.Height),
		zap.String("txHash", strings.ToUpper(hex.EncodeToString(txHash))),
		zap.Any("result", resp.TxResult))

	if resp.TxResult.Code != abci.CodeOk.Uint32() {
		return fmt.Errorf("transaction not ok, %s", resp.TxResult.Log)
	}
*/
