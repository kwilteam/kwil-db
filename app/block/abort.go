package block

import (
	"github.com/kwilteam/kwil-db/app/rpc"
	"github.com/kwilteam/kwil-db/app/shared/display"
	"github.com/spf13/cobra"
)

func abortCmd() *cobra.Command {
	var blockHeight int64
	var txIDs []string

	cmd := &cobra.Command{
		Use:   "abort <block_height> <tx_ids>",
		Short: "Aborts the execution of the current block if it's at block_height and removes the tx_ids from the mempool.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			clt, err := rpc.AdminSvcClient(ctx, cmd)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			err = clt.AbortBlockExecution(ctx, blockHeight, txIDs)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			return nil
		},
	}

	cmd.Flags().Int64VarP(&blockHeight, "block_height", "b", 0, "Block height to rollback")
	cmd.Flags().StringSliceVarP(&txIDs, "tx_ids", "t", nil, "Transaction IDs to remove from the mempool")
	return cmd
}
