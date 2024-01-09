package utils

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/spf13/cobra"
)

func queryTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query-tx",
		Short: "Query a transaction's status by hash.",
		Long:  "Query a transaction's status by hash. The hash should be passed as a hex-encoded string.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := common.GetAdminSvcClient(ctx, cmd)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			txHash, err := hex.DecodeString(args[0])
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			res, err := client.TxQuery(ctx, txHash)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			return display.PrintCmd(cmd, &display.RespTxQuery{Msg: res})
		},
	}

	common.BindRPCFlags(cmd)

	return cmd
}

// RespTxQuery is used to represent a transaction response in cli
type RespTxQuery struct {
	Msg *transactions.TcTxQueryResponse
}

func (r *RespTxQuery) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Hash     string                         `json:"hash"` // HEX
		Height   int64                          `json:"height"`
		Tx       transactions.Transaction       `json:"tx"`
		TxResult transactions.TransactionResult `json:"tx_result"`
	}{
		Hash:     hex.EncodeToString(r.Msg.Hash),
		Height:   r.Msg.Height,
		Tx:       r.Msg.Tx,
		TxResult: r.Msg.TxResult,
	})
}

func (r *RespTxQuery) MarshalText() ([]byte, error) {
	status := "failed"
	if r.Msg.Height == -1 {
		status = "pending"
	} else if r.Msg.TxResult.Code == transactions.CodeOk.Uint32() {
		status = "success"
	}

	msg := fmt.Sprintf(`Transaction ID: %s
Status: %s
Height: %d
Log: %s`,
		hex.EncodeToString(r.Msg.Hash),
		status,
		r.Msg.Height,
		r.Msg.TxResult.Log,
	)

	return []byte(msg), nil
}
