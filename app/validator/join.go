package validator

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/kwilteam/kwil-db/app/rpc"
	"github.com/kwilteam/kwil-db/app/shared/display"
)

var (
	joinLong = "A node may request to join the validator set by submitting a join request using the `join` command. The key used to sign the join request will be the treated as the node request to join the validator set. The node will be added to the validator set if the join request is approved by the current validator set. The status of a join request can be queried using the `join-status` command."

	joinExample = `# Request to join the validator set
kwil-admin validators join`
)

func joinCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "join",
		Short:   "A node may request to join the validator set by submitting a join request using the `join` command.",
		Long:    joinLong,
		Example: joinExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			clt, err := rpc.AdminSvcClient(ctx, cmd)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			txHash, err := clt.Join(ctx)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			return display.PrintCmd(cmd, display.RespTxHash(txHash))
		},
	}

	return cmd
}
