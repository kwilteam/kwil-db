package utils

import (
	"context"
	"encoding/hex"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/spf13/cobra"
)

func queryTxCmd() *cobra.Command {
	var raw bool
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

			return display.PrintCmd(cmd, &display.RespTxQuery{Msg: res, WithRaw: raw})
		},
	}

	cmd.Flags().BoolVarP(&raw, "raw", "R", false, "also display the bytes of the serialized transaction")

	common.BindRPCFlags(cmd)

	return cmd
}
